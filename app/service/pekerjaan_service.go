package service

import (
	"context"
	"time"

	"praktikummongo/app/model"
	"praktikummongo/app/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive" // <-- TAMBAHKAN IMPORT INI
)

type PekerjaanService struct {
	repo repository.IPekerjaanRepository
}

func NewPekerjaanService(repo repository.IPekerjaanRepository) *PekerjaanService {
	return &PekerjaanService{repo: repo}
}

// ------------------- CRUD Dasar -------------------

func (s *PekerjaanService) GetAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	list, err := s.repo.GetAll(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	return c.JSON(list)
}

func (s *PekerjaanService) GetByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	data, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	if data == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}
	return c.JSON(data)
}

func (s *PekerjaanService) GetByAlumniID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userVal := c.Locals("user_id")
	userID, ok := userVal.(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User ID tidak valid"})
	}

	list, err := s.repo.GetByAlumniUserID(ctx, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	return c.JSON(list)
}

func (s *PekerjaanService) Create(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var p model.PekerjaanAlumni
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid", "detail": err.Error()})
	}

	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	newData, err := s.repo.Create(ctx, &p)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menambah data", "detail": err.Error()})
	}
	return c.Status(201).JSON(newData)
}

func (s *PekerjaanService) Update(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	var p model.PekerjaanAlumni
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid", "detail": err.Error()})
	}

	p.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, id, &p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui data", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Pekerjaan berhasil diupdate"})
}

func (s *PekerjaanService) Delete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if err := s.repo.HardDelete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus data", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Pekerjaan berhasil dihapus"})
}

// ------------------- RBAC (Soft Delete, Restore, Hard Delete) -------------------

func (s *PekerjaanService) DeleteRBAC(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	// --- FIX 1 ---
	// Konversi userID (string) ke ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "User ID tidak valid"})
	}
	// --- END FIX 1 ---

	ownerID, err := s.repo.GetOwnerID(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	// --- FIX 2 ---
	// Bandingkan ObjectID dengan ObjectID, dan dereferensi ownerID
	if role != "admin" && *ownerID != userObjID {
		// --- END FIX 2 ---
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki izin menghapus data ini"})
	}

	if err := s.repo.SoftDelete(ctx, id, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal melakukan soft delete", "detail": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Pekerjaan berhasil dihapus (soft delete)"})
}

func (s *PekerjaanService) GetTrash(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	// --- FIX 3 ---
	// Konversi userID (string) ke *primitive.ObjectID
	// Hanya lakukan jika bukan admin, karena admin bisa melihat semua trash
	var userObjID *primitive.ObjectID
	if role != "admin" {
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "User ID tidak valid"})
		}
		userObjID = &objID // Kirim pointernya
	}
	// --- END FIX 3 ---

	// Kirim userObjID (yang sekarang bertipe *primitive.ObjectID)
	data, err := s.repo.GetTrash(ctx, role, userObjID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data", "detail": err.Error()})
	}
	if len(data) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Tidak ada pekerjaan yang dihapus"})
	}

	return c.JSON(data)
}

func (s *PekerjaanService) Restore(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	if err := s.repo.Restore(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal merestore data", "detail": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Pekerjaan berhasil direstore"})
}

func (s *PekerjaanService) HardDelete(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id := c.Params("id")
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("user_id").(string)

	// --- FIX 4 ---
	// Konversi userID (string) ke ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "User ID tidak valid"})
	}
	// --- END FIX 4 ---

	ownerID, deleted, err := s.repo.GetOwnerAndDeleteStatus(ctx, id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	// --- FIX 5 ---
	// Dereferensi pointer 'deleted'
	if deleted == nil || !*deleted {
		// --- END FIX 5 ---
		return c.Status(400).JSON(fiber.Map{"error": "Data belum dihapus (soft delete)"})
	}

	// --- FIX 6 ---
	// Bandingkan ObjectID dengan ObjectID
	if role != "admin" && *ownerID != userObjID {
		// --- END FIX 6 ---
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak memiliki izin menghapus permanen data ini"})
	}

	if err := s.repo.HardDelete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus permanen", "detail": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Pekerjaan berhasil dihapus permanen"})
}	