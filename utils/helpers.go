package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"

	"ec.com/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func IsAgency(token interface{}) bool {
	var u models.User
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	_ = json.Unmarshal([]byte(sub), &u)
	return u.Agency != "encerrar"
}

func GetUserID(token interface{}) uuid.UUID {
	var u models.User
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	if err := json.Unmarshal([]byte(sub), &u); err != nil {
		return uuid.Nil
	}
	return u.ID
}

func GenerateAuthCode() string {
	// Generate 16 random bytes
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}

	// Hash the random bytes with MD5
	hash := md5.Sum(b)

	// Convert to hex string
	return hex.EncodeToString(hash[:])
}

func GetAgencyId(token interface{}) uuid.UUID {
	var u models.User
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	_ = json.Unmarshal([]byte(sub), &u)
	return u.ID
}

func GetAgency(token interface{}) string {
	var u models.User
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	if err := json.Unmarshal([]byte(sub), &u); err != nil {
		return ""
	}
	return u.Agency
}
