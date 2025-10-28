package repository

import (
	"context"
	"errors"
	"praktikummongo/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // <-- PASTIKAN IMPORT INI
	"go.mongodb.org/mongo-driver/mongo"
)

type IUserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
}

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) IUserRepository {
	return &UserRepository{collection: db.Collection("users")}
}

// Ambil user berdasarkan username
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Jika tidak ada dokumen, kembalikan (nil, nil) -> BUKAN error
			return nil, nil
		}
		// Error database lainnya
		return nil, err
	}
	return &user, nil
}

// Tambah user baru
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	// Cek apakah username sudah dipakai
	var existing model.User
	err := r.collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existing)
	if err == nil {
		return nil, errors.New("username sudah digunakan")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err // Error database
	}

	// Simpan user baru
	res, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	// --- PERBAIKAN LOGIKA ---
	// res.InsertedID adalah primitive.ObjectID, bukan model.User
	user.ID = res.InsertedID.(primitive.ObjectID)
	// --- END PERBAIKAN ---

	return user, nil
}