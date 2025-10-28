package model

import (
    "github.com/golang-jwt/jwt/v5"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Username string             `bson:"username" json:"username"`
    Password string             `bson:"password" json:"password"` // hashed
    Role     string             `bson:"role" json:"role"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    User  User   `json:"user"`
    Token string `json:"token"`
}

type JWTClaims struct {
    UserID   primitive.ObjectID `json:"user_id"`
    Username string             `json:"username"`
    Role     string             `json:"role"`
    jwt.RegisteredClaims
}
