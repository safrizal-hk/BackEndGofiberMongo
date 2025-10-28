package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// File adalah model untuk data di MongoDB
type File struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FileName     string             `json:"file_name" bson:"file_name"`
	OriginalName string             `json:"original_name" bson:"original_name"`
	FilePath     string             `json:"file_path" bson:"file_path"`
	FileSize     int64              `json:"file_size" bson:"file_size"`
	FileType     string             `json:"file_type" bson:"file_type"`
	UploadedAt   time.Time          `json:"uploaded_at" bson:"uploaded_at"`
}

// FileResponse adalah model untuk response JSON
type FileResponse struct {
	ID           string    `json:"id"`
	FileName     string    `json:"file_name"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	FileType     string    `json:"file_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
}