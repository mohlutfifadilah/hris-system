package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// Index - Tampilkan halaman dashboard
func (dc *DashboardController) Index(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	// Hitung total karyawan
	var totalEmployees int64
	config.DB.Model(&models.Employee{}).Count(&totalEmployees)

	// Render dashboard menggunakan layout main.html
	c.HTML(http.StatusOK, "main.html", gin.H{
		"title":          "Dashboard",
		"user":           currentUser, // seluruh row employee yang login (boleh nil)
		"activePage":     "dashboard",
		"totalEmployees": totalEmployees,
	})
}
