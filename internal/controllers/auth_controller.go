package controllers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"hris-system/config"
	"hris-system/internal/middleware"
	"hris-system/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// ============================================
// ShowLoginForm - Tampilkan halaman login
// ============================================
func (ac *AuthController) ShowLoginForm(c *gin.Context) {
	flashMessage, _ := c.Cookie("flash_message")
	if flashMessage != "" {
		c.SetCookie("flash_message", "", -1, "/", "", false, false)
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title":        "Login",
		"flashMessage": flashMessage,
	})
}

// ============================================
// Login - Proses autentikasi user
// ============================================
func (ac *AuthController) Login(c *gin.Context) {
	// 1. Parse request body
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email dan password harus diisi",
		})
		return
	}

	// 2. Normalize & validate email
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Format email tidak valid",
			"field":   "email",
		})
		return
	}

	// 3. Validate password length
	if len(input.Password) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Password minimal 3 karakter",
			"field":   "password",
		})
		return
	}

	// 4. Check email exists in database
	var employee models.Employee
	if err := config.DB.Where("email = ?", email).First(&employee).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Email atau password salah",
			"field":   "email",
		})
		return
	}

	// 5. Verify password with pgcrypto
	if !verifyPassword(email, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Email atau password salah",
			"field":   "password",
		})
		return
	}

	// 6. Generate JWT token & set cookie
	tokenString, err := generateToken(employee.Email, employee.Nama)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal membuat token",
		})
		return
	}

	c.SetCookie("token", tokenString, 3600, "/", "", false, true) // 24 hours

	// 7. Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login berhasil!",
		"user": gin.H{
			"nama":  employee.Nama,
			"email": employee.Email,
			"unit":  employee.Unit,
		},
	})
}

// ============================================
// Logout - Hapus session user
// ============================================
func (ac *AuthController) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	// c.SetCookie("flash_message", "Anda telah berhasil logout", 5, "/", "", false, false)
	c.Redirect(http.StatusSeeOther, "/")
}

// ============================================
// HELPER FUNCTIONS (Private)
// ============================================

// isValidEmail - Cek format email
func isValidEmail(email string) bool {
	if email == "" || len(email) < 5 || len(email) > 100 {
		return false
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// verifyPassword - Verifikasi password dengan pgcrypto
func verifyPassword(email, password string) bool {
	var match bool
	query := "SELECT password = crypt(?, password) FROM employees WHERE email = ?"
	config.DB.Raw(query, password, email).Scan(&match)
	return match
}

// generateToken - Generate JWT token
func generateToken(email, nama string) (string, error) {
	claims := &middleware.Claims{
		Email: email,
		Nama:  nama,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(middleware.JWTSecret)
}
