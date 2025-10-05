package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var jwtSecret = []byte(getenv("JWT_SECRET", "supersecretkey"))

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func CreateToken(userID primitive.ObjectID, ttlHours int) (string, error) {
	if ttlHours <= 0 {
		ttlHours = 72
	}
	claims := jwt.MapClaims{
		"sub": userID.Hex(),
		"exp": time.Now().Add(time.Duration(ttlHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenStr string) (primitive.ObjectID, error) {
	if tokenStr == "" {
		return primitive.NilObjectID, errors.New("empty token")
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// ensure signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return primitive.NilObjectID, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, errors.New("invalid token claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return primitive.NilObjectID, errors.New("missing sub claim")
	}
	id, err := primitive.ObjectIDFromHex(sub)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return id, nil
}
