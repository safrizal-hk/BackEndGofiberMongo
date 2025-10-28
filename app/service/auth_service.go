package service

import (
	"context"
	"time"

	"praktikummongo/app/model"
	"praktikummongo/app/repository"
	"praktikummongo/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	repo repository.IUserRepository
}

func NewAuthService(repo repository.IUserRepository) *AuthService {
	return &AuthService{repo: repo}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ---------------------- LOGIN ----------------------

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ambil user dari MongoDB berdasarkan username
	// --- PERBAIKAN 1: Nama method salah ---
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	// --- END PERBAIKAN 1 ---

	// --- PERBAIKAN 2: Error handling logika ---
	// Cek error database dulu
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Terjadi kesalahan server", "detail": err.Error()})
	}
	// Cek jika user tidak ditemukan (user == nil)
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "User tidak ditemukan"})
	}
	// --- END PERBAIKAN 2 ---

	// Validasi password (hash dan fallback plaintext)
	if utils.CheckPasswordHash(req.Password, user.Password) {
		// valid hash
	} else if req.Password == user.Password {
		// fallback plaintext
	} else {
		return c.Status(401).JSON(fiber.Map{"error": "Password salah"})
	}

	// Generate token JWT
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// ---------------------- REGISTER ----------------------

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.User
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Cek apakah username sudah digunakan
	// --- PERBAIKAN 3: Nama method salah & error check ---
	existing, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Terjadi kesalahan server", "detail": err.Error()})
	}
	// --- END PERBAIKAN 3 ---
	if existing != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Username sudah terdaftar"})
	}

	// Hash password
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengenkripsi password"})
	}
	req.Password = hashed
	req.Role = "user"

	// --- PERBAIKAN 4: Field tidak ada di model.User ---
	// req.CreatedAt = time.Now() // <-- HAPUS
	// req.UpdatedAt = time.Now() // <-- HAPUS
	// --- END PERBAIKAN 4 ---

	// Simpan user baru ke MongoDB
	// --- PERBAIKAN 5: Nama method salah ---
	newUser, err := s.repo.CreateUser(ctx, &req)
	// --- END PERBAIKAN 5 ---
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan user", "detail": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Registrasi berhasil",
		"user": fiber.Map{
			"id":       newUser.ID.Hex(),
			"username": newUser.Username,
			"role":     newUser.Role,
		},
	})
}