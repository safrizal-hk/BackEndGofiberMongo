package repository

import (
	"context"
	"errors"
	"praktikummongo/app/model"
	"time" // <-- PASTIKAN IMPORT TIME

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IPekerjaanRepository interface {
	GetAll(ctx context.Context) ([]model.PekerjaanAlumni, error)
	GetByID(ctx context.Context, id string) (*model.PekerjaanAlumni, error)
	GetByAlumniID(ctx context.Context, alumniID string) ([]model.PekerjaanAlumni, error)
	GetByAlumniUserID(ctx context.Context, userID string) ([]model.PekerjaanAlumni, error)
	Create(ctx context.Context, pekerjaan *model.PekerjaanAlumni) (*model.PekerjaanAlumni, error)
	Update(ctx context.Context, id string, pekerjaan *model.PekerjaanAlumni) error
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string, userID string) error
	GetTrash(ctx context.Context, role string, userID *primitive.ObjectID) ([]model.TrashPekerjaan, error)
	Restore(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	GetOwnerID(ctx context.Context, pekerjaanID string) (*primitive.ObjectID, error)
	GetOwnerAndDeleteStatus(ctx context.Context, id string) (*primitive.ObjectID, *bool, error)
}

type PekerjaanRepository struct {
	collection *mongo.Collection
	alumniColl *mongo.Collection
}

func NewPekerjaanRepository(db *mongo.Database) IPekerjaanRepository {
	return &PekerjaanRepository{
		collection: db.Collection("pekerjaan_alumni"),
		alumniColl: db.Collection("alumni"),
	}
}

// Ambil semua pekerjaan aktif
// (FIXED: Mencari yang is_deleted TIDAK ADA atau NIL)
func (r *PekerjaanRepository) GetAll(ctx context.Context) ([]model.PekerjaanAlumni, error) {
	// Temukan dokumen di mana 'is_deleted' tidak ada (null)
	cursor, err := r.collection.Find(ctx, bson.M{"is_deleted": nil})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []model.PekerjaanAlumni
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Ambil pekerjaan berdasarkan ID
func (r *PekerjaanRepository) GetByID(ctx context.Context, id string) (*model.PekerjaanAlumni, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	var pekerjaan model.PekerjaanAlumni
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &pekerjaan, nil
}

// ... (Fungsi GetByAlumniID, GetByAlumniUserID, Create, Update, Delete tetap sama) ...

// Ambil pekerjaan berdasarkan alumni_id
func (r *PekerjaanRepository) GetByAlumniID(ctx context.Context, alumniID string) ([]model.PekerjaanAlumni, error) {
	alumniObjID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, errors.New("Alumni ID tidak valid")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"alumni_id": alumniObjID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []model.PekerjaanAlumni
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Ambil pekerjaan berdasarkan userID
func (r *PekerjaanRepository) GetByAlumniUserID(ctx context.Context, userID string) ([]model.PekerjaanAlumni, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("User ID tidak valid")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []model.PekerjaanAlumni
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Tambah pekerjaan baru
func (r *PekerjaanRepository) Create(ctx context.Context, pekerjaan *model.PekerjaanAlumni) (*model.PekerjaanAlumni, error) {
	res, err := r.collection.InsertOne(ctx, pekerjaan)
	if err != nil {
		return nil, err
	}
	pekerjaan.ID = res.InsertedID.(primitive.ObjectID)
	return pekerjaan, nil
}

// Update pekerjaan
func (r *PekerjaanRepository) Update(ctx context.Context, id string, pekerjaan *model.PekerjaanAlumni) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	update := bson.M{"$set": pekerjaan}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Hapus permanen
func (r *PekerjaanRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Soft delete dengan userID
// (FIXED: Mengisi is_deleted dengan time.Now())
func (r *PekerjaanRepository) SoftDelete(ctx context.Context, id string, userID string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": time.Now(), // <-- Diubah menjadi timestamp
			"deleted_by": userID,
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Ambil daftar pekerjaan yang sudah dihapus
// (FIXED: Mencari yang is_deleted ADA / NOT NIL)
func (r *PekerjaanRepository) GetTrash(ctx context.Context, role string, userID *primitive.ObjectID) ([]model.TrashPekerjaan, error) {
	// Temukan dokumen di mana 'is_deleted' ada (bukan null)
	filter := bson.M{"is_deleted": bson.M{"$ne": nil}}

	if role != "admin" && userID != nil {
		filter["alumni_id"] = *userID
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Model TrashPekerjaan Anda sudah benar menggunakan *time.Time
	var trashList []model.TrashPekerjaan
	if err := cursor.All(ctx, &trashList); err != nil {
		return nil, err
	}
	return trashList, nil
}

// Restore pekerjaan
// (Logic $unset sudah benar untuk menghapus field timestamp)
func (r *PekerjaanRepository) Restore(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	// $unset akan menghapus field 'is_deleted', membuatnya jadi nil (aktif kembali)
	update := bson.M{"$unset": bson.M{"is_deleted": "", "deleted_by": ""}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Hapus permanen
func (r *PekerjaanRepository) HardDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("ID tidak valid")
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// Ambil pemilik pekerjaan
func (r *PekerjaanRepository) GetOwnerID(ctx context.Context, pekerjaanID string) (*primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(pekerjaanID)
	if err != nil {
		return nil, errors.New("ID tidak valid")
	}

	var pekerjaan model.PekerjaanAlumni
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan)
	if err != nil {
		return nil, err
	}
	// Asumsi model.PekerjaanAlumni punya AlumniID
	return &pekerjaan.AlumniID, nil 
}

// Ambil pemilik dan status delete
// (FIXED: Mengecek apakah is_deleted ada/nil, bukan true/false)
func (r *PekerjaanRepository) GetOwnerAndDeleteStatus(ctx context.Context, id string) (*primitive.ObjectID, *bool, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil, errors.New("ID tidak valid")
	}

	// Gunakan bson.M agar fleksibel
	var result bson.M
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		return nil, nil, err
	}

	// 1. Ambil Owner ID
	var ownerID primitive.ObjectID
	if oid, ok := result["alumni_id"].(primitive.ObjectID); ok {
		ownerID = oid
	} else {
		return nil, nil, errors.New("field 'alumni_id' tidak ditemukan atau tipe salah")
	}

	// 2. Ambil Deletion Status (FIXED)
	isDeleted := false // Default-nya false (tidak terhapus)
	
	// Cek apakah field 'is_deleted' ada dan BUKAN nil
	if val, ok := result["is_deleted"]; ok && val != nil {
		isDeleted = true // Jika ada isinya (timestamp), berarti terhapus
	}

	return &ownerID, &isDeleted, nil
}