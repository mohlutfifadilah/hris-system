package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "default-jwt-secret-key-please-set-env"
	}
	return secret
}

type Claims struct {
	Email string `json:"email"`
	Nama  string `json:"nama"`
	jwt.RegisteredClaims
}

// AuthRequired - Middleware untuk protect routes
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie("token")
		if err != nil {
			// Set session untuk flash message
			c.SetCookie("flash_message", "Silakan login terlebih dahulu", 5, "/", "", false, false)
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.SetCookie("flash_message", "Sesi Anda telah berakhir. Silakan login kembali", 5, "/", "", false, false)
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// Set user info to context
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("userEmail", claims.Email)
			c.Set("userName", claims.Nama)
		}

		c.Next()
	}
}

// RedirectIfAuthenticated - Redirect ke dashboard jika sudah login
func RedirectIfAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err == nil {
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return JWTSecret, nil
			})

			if err == nil && token.Valid {
				// Sudah login, redirect ke employees
				c.Redirect(http.StatusSeeOther, "/dashboard")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
