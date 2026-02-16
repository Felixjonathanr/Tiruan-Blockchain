package auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(nama string, pvKey []byte) (string, error) {
	claims := jwt.MapClaims{
		"user_nama": nama,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(pvKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func Middleware(kunciRahasia []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return kunciRahasia, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{
				"error": "kamu tidak memiliki akses",
			})
			c.Abort()

			return
		}
		c.Next()
	}
}
