package model

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type TrashPekerjaan struct {
    ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
    AlumniID       primitive.ObjectID  `bson:"alumni_id" json:"alumni_id"`
    UserID         *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
    NamaAlumni     string              `bson:"nama_alumni" json:"nama_alumni"`
    NamaPerusahaan string              `bson:"nama_perusahaan" json:"nama_perusahaan"`
    PosisiJabatan  string              `bson:"posisi_jabatan" json:"posisi_jabatan"`
    BidangIndustri string              `bson:"bidang_industri" json:"bidang_industri"`
    LokasiKerja    string              `bson:"lokasi_kerja" json:"lokasi_kerja"`
    IsDeleted      *time.Time          `bson:"is_deleted,omitempty" json:"is_deleted,omitempty"`
}
