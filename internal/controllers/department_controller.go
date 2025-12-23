package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-gonic/gin"
)

type DepartmentController struct{}

func NewDepartmentController() *DepartmentController {
	return &DepartmentController{}
}

// Index - Tampilkan halaman profile
func (dc *DepartmentController) Index(c *gin.Context) {
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

	// // 2) Ambil row career lengkap dari DB
	// var career models.Career
	// if err := config.DB.
	// 	Where("id = ?", employee.IDCareer).
	// 	First(&career).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "career not found")
	// 	return
	// }

	// // 3) Ambil row career_history lengkap dari DB
	// var career_history models.CareerHistory
	// if err := config.DB.
	// 	Where("id = ?", career.IDCareerHistory).
	// 	First(&career_history).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "career history not found")
	// 	return
	// }

	// // 4) Ambil row department_history lengkap dari DB
	// var department_history models.DepartmentHistory
	// if err := config.DB.
	// 	Where("id = ?", career_history.IDDepartment).
	// 	First(&department_history).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "department history not found")
	// 	return
	// }

	// Render department menggunakan layout main.html
	c.HTML(http.StatusOK, "department", gin.H{
		"title": "Department",
		"user":  employee, // seluruh row employee yang login (boleh nil)
		// "department": department_history, // seluruh row employee yang login (boleh nil)
		"activePage": "department",
	})
}
