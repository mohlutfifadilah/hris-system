package controllers

import (
	"net/http"

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
