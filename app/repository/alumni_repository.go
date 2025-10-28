package repository

import (
	"context"
	"errors"
	"praktikummongo/app/model" // Pastikan model di-import

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IAlumniRepository interface {
	GetAll(ctx context.Context) ([]model.Alumni, error)
	GetByID(ctx context.Context, id string) (*model.Alumni, error)
	Create(ctx context.Context, alumni *model.Alumni) (*model.Alumni, error)
	Update(ctx context.Context, id string, alumni *model.Alumni) error
	Delete(ctx context.Context, id string) error
	GetWithFilter(ctx context.Context, page, limit int, sortBy, order, search string) ([]model.Alumni, int, error)
	// --- TAMBAHKAN METHOD INI KE INTERFACE ---
	GetJumlahByAngkatan(ctx context.Context) ([]model.JumlahAngkatan, error)
}

type AlumniRepository struct {
	collection *mongo.Collection
}

func NewAlumniRepository(db *mongo.Database) IAlumniRepository {
	return &AlumniRepository{collection: db.Collection("alumni")}
}

// Ambil semua data alumni
func (r *AlumniRepository) GetAll(ctx context.Context) ([]model.Alumni, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Alumni
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Ambil alumni berdasarkan ID
func (r *AlumniRepository) GetByID(ctx context.Context, id string) (*model.Alumni, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	var alumni model.Alumni
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&alumni)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &alumni, nil
}

// Tambah alumni baru
func (r *AlumniRepository) Create(ctx context.Context, alumni *model.Alumni) (*model.Alumni, error) {
	alumni.ID = primitive.NilObjectID
	result, err := r.collection.InsertOne(ctx, alumni)
	if err != nil {
		return nil, err
	}
	alumni.ID = result.InsertedID.(primitive.ObjectID)
	return alumni, nil
}

// Update data alumni
func (r *AlumniRepository) Update(ctx context.Context, id string, alumni *model.Alumni) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}

	update := bson.M{
		"$set": bson.M{
			"nama":        alumni.Nama,
			"jurusan":     alumni.Jurusan,
			"angkatan":    alumni.Angkatan,
			"tahun_lulus": alumni.TahunLulus,
			"no_telepon":  alumni.NoTelepon,
			"alamat":      alumni.Alamat,
			"updated_at":  alumni.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Hapus alumni permanen
func (r *AlumniRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// GetWithFilter - Mendapatkan data alumni dengan pagination, sorting, dan search
func (r *AlumniRepository) GetWithFilter(ctx context.Context, page, limit int, sortBy, order, search string) ([]model.Alumni, int, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	skip := (page - 1) * limit

	// Filter pencarian
	filter := bson.M{}
	if search != "" {
		filter["$or"] = []bson.M{
			{"nama": bson.M{"$regex": search, "$options": "i"}},
			{"jurusan": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	// Sorting
	sortOrder := 1
	if order == "desc" || order == "DESC" {
		sortOrder = -1
	}
	sortStage := bson.D{{Key: sortBy, Value: sortOrder}}

	// Pipeline agregasi untuk pagination
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: sortStage}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var result []model.Alumni
	if err := cursor.All(ctx, &result); err != nil {
		return nil, 0, err
	}

	// Hitung total dokumen
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return result, int(total), nil
}

// --- TAMBAHKAN IMPLEMENTASI FUNGSI INI ---
// GetJumlahByAngkatan - Menjalankan agregasi untuk menghitung jumlah alumni per angkatan
func (r *AlumniRepository) GetJumlahByAngkatan(ctx context.Context) ([]model.JumlahAngkatan, error) {
	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: "$angkatan"},
		{Key: "jumlah", Value: bson.D{{Key: "$sum", Value: 1}}},
	}}}
	sortStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}}

	// Sekarang kita bisa mengakses r.collection karena berada di package yang sama
	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{groupStage, sortStage})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Gunakan struct model.JumlahAngkatan yang sudah kita buat
	var results []model.JumlahAngkatan
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}