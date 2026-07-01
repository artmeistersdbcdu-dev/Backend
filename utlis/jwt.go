package utlis

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var JWTSecert []byte = []byte(config.GetConfig().JWTSecert)

func GenerateJwt(userId uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"userId": userId.String(),
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecert)

}
func GetUserIdJwt(cookie *http.Cookie) (string, error) {
	tokenString := cookie.Value
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return JWTSecert, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	userID, ok := claims["userId"].(string)
	if !ok {
		return "", err
	}

	return userID, nil
}
