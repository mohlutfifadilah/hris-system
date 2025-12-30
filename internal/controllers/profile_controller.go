package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-gonic/gin"
)

type ProfileController struct{}

func NewProfileController() *ProfileController {
	return &ProfileController{}
}

// Index - Tampilkan halaman profile
func (dc *ProfileController) Index(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	// 1) Ambil row employee lengkap dari DB
	var employee models.Employee
	if err := config.DB.
		Where("id = ?", currentUser.ID).
		First(&employee).Error; err != nil {
		// handle error (404, dll)
		c.String(http.StatusInternalServerError, "employee not found")
		return
	}

	// Render profile menggunakan layout main.html
	c.HTML(http.StatusOK, "profile", gin.H{
		"title":      "Profile",
		"user":       employee, // seluruh row employee yang login (boleh nil)
		"activePage": "profile",
	})
}

// UploadProfilePhoto untuk upload foto profile
func (dc *ProfileController) UploadProfilePhoto(c *gin.Context) {
	userID := auth.GetCurrentUser(c)

	// Ambil file dari form
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
		return
	}

	// Validasi ekstensi file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file not relevant. Use JPG or PNG"})
		return
	}

	// Validasi ukuran file (max 2MB)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size max 2MB"})
		return
	}

	// Get employee data
	var user models.Employee
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hapus foto lama jika ada
	if user.Photo != "" && user.Photo != "static/assets/img/avatar/avatar-1.png" {
		oldPhotoPath := user.Photo
		if _, err := os.Stat(oldPhotoPath); err == nil {
			os.Remove(oldPhotoPath)
		}
	}

	// Generate nama file unik
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("profile_%d_%d%s", userID, timestamp, ext)

	// Path untuk menyimpan file
	uploadDir := "static/assets/img/profiles"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Directory"})
		return
	}

	filepath := filepath.Join(uploadDir, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Upload File"})
		return
	}

	// Update database
	user.Photo = filepath
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Update Database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto Profile success upload",
		"photo":   filepath,
	})
}

// UpdateProfilePhoto untuk edit foto profile
func (dc *ProfileController) UpdateProfilePhoto(c *gin.Context) {
	// Sama seperti UploadProfilePhoto, karena fungsinya mengganti foto
	dc.UploadProfilePhoto(c)
}

// DeleteProfilePhoto untuk hapus foto profile
func (dc *ProfileController) DeleteProfilePhoto(c *gin.Context) {
	userID := auth.GetCurrentUser(c)

	var user models.Employee
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Hapus file foto jika ada
	if user.Photo != "" && user.Photo != "static/assets/img/avatar/avatar-1.png" {
		if _, err := os.Stat(user.Photo); err == nil {
			if err := os.Remove(user.Photo); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file"})
				return
			}
		}
	}

	// Set ke default avatar
	user.Photo = "static/assets/img/avatar/avatar-1.png"
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error update database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto profile success deleted",
		"photo":   user.Photo,
	})
}
