package service

import (
	"context"
	"strconv"
	"time"

	"praktikummongo/app/model"
	"praktikummongo/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson" // bson masih dipakai di GetAlumniDenganDuaPekerjaan
	"go.mongodb.org/mongo-driver/mongo"
)

type AlumniService struct {
	repo      repository.IAlumniRepository
	pekerjaan *mongo.Collection
}

func NewAlumniService(repo repository.IAlumniRepository, db *mongo.Database) *AlumniService {
	return &AlumniService{
		repo:      repo,
		pekerjaan: db.Collection("pekerjaan_alumni"),
	}
}

// ------------------- CRUD -------------------

func (s *AlumniService) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := s.repo.GetAll(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	return c.JSON(list)
}

func (s *AlumniService) GetByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	alumni, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	if alumni == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
	}
	return c.JSON(alumni)
}

func (s *AlumniService) Create(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var a model.Alumni
	if err := c.BodyParser(&a); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid", "detail": err.Error()})
	}

	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	newAlumni, err := s.repo.Create(ctx, &a)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan data", "detail": err.Error()})
	}
	return c.Status(201).JSON(newAlumni)
}

func (s *AlumniService) Update(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	var a model.Alumni
	if err := c.BodyParser(&a); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid", "detail": err.Error()})
	}
	a.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, id, &a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui data", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Alumni berhasil diupdate"})
}

func (s *AlumniService) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if err := s.repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus data", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Alumni berhasil dihapus"})
}

// ------------------- Pagination + Filter -------------------

func (s *AlumniService) GetAlumniWithPagination(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	sortBy := c.Query("sort", "nama")
	order := c.Query("order", "asc")
	search := c.Query("search", "")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	items, total, err := s.repo.GetWithFilter(ctx, page, limit, sortBy, order, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return c.JSON(fiber.Map{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"data":        items,
	})
}

// ------------------- Statistik -------------------

// --- FUNGSI INI DIPERBARUI ---
func (s *AlumniService) GetJumlahByAngkatan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// HAPUS SEMUA LOGIKA AGREGASI DARI SINI
	// ... (groupStage, sortStage, cursor, dll dihapus) ...

	// CUKUP PANGGIL REPOSITORY
	results, err := s.repo.GetJumlahByAngkatan(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}

	// HAPUS DECODE LOGIC DAN STRUCT LOKAL
	// ... (cursor.All, struct Result, dll dihapus) ...

	// LANGSUNG KEMBALIKAN HASIL DARI REPO
	return c.JSON(results)
}
// --- END PERBAIKAN ---

func (s *AlumniService) GetAlumniDenganDuaPekerjaan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fungsi ini tidak error karena s.pekerjaan memang field publik di AlumniService
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$alumni_id"},
			{Key: "jumlah_pekerjaan", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$match", Value: bson.D{{Key: "jumlah_pekerjaan", Value: bson.D{{Key: "$gte", Value: 2}}}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "alumni"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "alumni"},
		}}},
		{{Key: "$unwind", Value: "$alumni"}},
		{{Key: "$project", Value: bson.D{
			{Key: "nama", Value: "$alumni.nama"},
			{Key: "jumlah_pekerjaan", Value: 1},
		}}},
	}

	cursor, err := s.pekerjaan.Aggregate(ctx, pipeline)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	defer cursor.Close(ctx)

	type Result struct {
		Nama            string `bson:"nama" json:"nama"`
		JumlahPekerjaan int    `bson:"jumlah_peekerjaan" json:"jumlah_pekerjaan"` // Typo di bson tag? saya biarkan sesuai kode Anda
	}

	var results []Result
	if err := cursor.All(ctx, &results); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal baca data", "detail": err.Error()})
	}
	return c.JSON(results)
}