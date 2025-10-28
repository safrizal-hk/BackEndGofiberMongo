package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB() (*mongo.Client, *mongo.Database) {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")

	if mongoURI == "" {
		log.Fatal("MONGO_URI belum diatur di .env")
	}
	if dbName == "" {
		log.Fatal("MONGO_DB belum diatur di .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Gagal konek MongoDB:", err)
	}

	db := client.Database(dbName)
	log.Println("Berhasil terhubung ke MongoDB:", dbName)
	return client, db
}
