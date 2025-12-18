package controllers

import (
    "net/http"
    "regexp"
    "strings"
    "time"
    
    "hris-system/config"
    "hris-system/models"
    "hris-system/internal/middleware"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type AuthController struct{}

func NewAuthController() *AuthController {
    return &AuthController{}
}

// ShowLoginForm - Tampilkan halaman login
func (ac *AuthController) ShowLoginForm(c *gin.Context) {
    // Get flash message dari cookie
    flashMessage, _ := c.Cookie("flash_message")
    
    // Hapus flash message cookie setelah dibaca
    if flashMessage != "" {
        c.SetCookie("flash_message", "", -1, "/", "", false, false)
    }
    
    c.HTML(http.StatusOK, "login.html", gin.H{
        "title":        "Login",
        "flashMessage": flashMessage,
    })
}

// ValidateEmail - Validasi format email
func (ac *AuthController) ValidateEmail(email string) (bool, string) {
    // Trim whitespace
    email = strings.TrimSpace(email)
    
    // Cek kosong
    if email == "" {
        return false, "Email tidak boleh kosong"
    }
    
    // Cek format email
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return false, "Format email tidak valid"
	}
    
    return true, ""
}

// ValidatePassword - Validasi password
func (ac *AuthController) ValidatePassword(password string) (bool, string) {
    // Cek kosong
    if password == "" {
        return false, "Password tidak boleh kosong"
    }
    
    // Cek panjang minimal
    if len(password) < 3 {
        return false, "Password minimal 3 karakter"
    }
    
    return true, ""
}

// Login - Proses login dengan validasi ketat
func (ac *AuthController) Login(c *gin.Context) {
    var input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Data tidak valid",
        })
        return
    }

    // Validasi Email
    validEmail, errMsgEmail := ac.ValidateEmail(input.Email)
    if !validEmail {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": errMsgEmail,
            "field":   "email",
        })
        return
    }

    // Validasi Password
    validPassword, errMsgPassword := ac.ValidatePassword(input.Password)
    if !validPassword {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": errMsgPassword,
            "field":   "password",
        })
        return
    }

    // Trim dan lowercase email
    input.Email = strings.ToLower(strings.TrimSpace(input.Email))

    // Cari user di database
    var employee models.Employee
    if err := config.DB.Where("email = ?", input.Email).First(&employee).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "Email atau password salah",
            "field":   "general",
        })
        return
    }

    // Verifikasi password menggunakan pgcrypto
    var passwordMatch bool
    query := "SELECT password = crypt(?, password) FROM employees WHERE email = ?"
    config.DB.Raw(query, input.Password, input.Email).Scan(&passwordMatch)

    if !passwordMatch {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "Email atau password salah",
            "field":   "password",
        })
        return
    }

    // Generate JWT token
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &middleware.Claims{
        Email: employee.Email,
        Nama:  employee.Nama,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(middleware.JWTSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Gagal membuat token",
            "field":   "general",
        })
        return
    }

    // Set cookie dengan secure flags
    c.SetCookie(
        "token",
        tokenString,
        int(24*time.Hour.Seconds()),
        "/",
        "",
        false, // secure: true di production dengan HTTPS
        true,  // httpOnly: true untuk keamanan
    )

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

// Logout - Proses logout
func (ac *AuthController) Logout(c *gin.Context) {
    // Hapus cookie token
    c.SetCookie("token", "", -1, "/", "", false, true)
    
    // Set flash message
    c.SetCookie("flash_message", "Anda telah berhasil logout", 5, "/", "", false, false)
    
    c.Redirect(http.StatusSeeOther, "/login")
}
