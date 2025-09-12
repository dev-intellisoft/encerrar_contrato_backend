package utils

import (
	"crypto/md5"
	"crypto/rand"
	"ec.com/models"
	"encoding/hex"
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"io"
)

func GetUserID(token interface{}) int64 {
	var u models.User
	user := token.(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	sub := claims["sub"].(string)
	if err := json.Unmarshal([]byte(sub), &u); err != nil {
		return 0
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
