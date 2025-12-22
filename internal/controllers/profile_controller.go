package controllers

import (
	"net/http"

	auth "hris-system/internal/auth"

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

	// Render profile menggunakan layout main.html
	c.HTML(http.StatusOK, "profile", gin.H{
		"title":      "Profile",
		"user":       currentUser, // seluruh row employee yang login (boleh nil)
		"activePage": "profile",
	})
}
