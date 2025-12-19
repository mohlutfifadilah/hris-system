package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// Index - Tampilkan halaman dashboard
func (dc *DashboardController) Index(c *gin.Context) {
	// Ambil user info dari context (di-set oleh middleware setelah login)
	// Untuk development, pakai dummy data jika belum ada session
	// userName, exists := c.Get("userName")
	// if !exists {
	// 	userName = "Guest User" // Default jika belum login
	// }

	// userEmail, exists := c.Get("userEmail")
	// if !exists {
	// 	userEmail = "guest@hris.com" // Default jika belum login
	// }

	// // Hitung total karyawan
	// var totalEmployees int64
	// config.DB.Model(&models.Employee{}).Count(&totalEmployees)

	// // Hitung karyawan per unit
	// var unitStats []struct {
	// 	Unit  string
	// 	Count int64
	// }
	// config.DB.Model(&models.Employee{}).
	// 	Select("unit, COUNT(*) as count").
	// 	Group("unit").
	// 	Order("count DESC").
	// 	Scan(&unitStats)

	// Render dashboard
	c.HTML(http.StatusOK, "main.html", gin.H{
		"title": "Dashboard",
		// "userName":       userName,
		// "userEmail":      userEmail,
		"activePage": "dashboard",
		// "totalEmployees": totalEmployees,
		// "unitStats":      unitStats,
	})
}
