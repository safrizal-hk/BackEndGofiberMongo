package model

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Alumni struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    NIM        string             `bson:"nim" json:"nim"`
    Nama       string             `bson:"nama" json:"nama"`
    Jurusan    string             `bson:"jurusan" json:"jurusan"`
    Angkatan   int                `bson:"angkatan" json:"angkatan"`
    TahunLulus int                `bson:"tahun_lulus" json:"tahun_lulus"`
    Email      string             `bson:"email" json:"email"`
    NoTelepon  string             `bson:"no_telepon" json:"no_telepon"`
    Alamat     string             `bson:"alamat" json:"alamat"`
    CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type JumlahAngkatan struct {
	Angkatan int `bson:"_id" json:"angkatan"`
	Jumlah   int `bson:"jumlah" json:"jumlah"`
}