package main

import (
	"context"
	"log"
	"os"

	"praktikummongo/database"
	"praktikummongo/route"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Gagal load file .env")
	}

	// Koneksi MongoDB
	client, db := database.ConnectMongoDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal("Gagal disconnect MongoDB:", err)
		}
	}()

	// Inisialisasi Fiber app dan inject DB
	// Pemanggilan 'route.NewApp(db)' sekarang sudah benar
	app := config.NewApp(db)

	// Ambil port dari .env
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	srv := ":" + port
	log.Println("Server berjalan di http://localhost" + srv)
	log.Fatal(app.Listen(srv))
}