package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type StatusController struct{}

func NewStatusController() *StatusController {
	return &StatusController{}
}

// Index - Tampilkan halaman
func (dc *StatusController) Index(c *gin.Context) {
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

	var status []models.Status
	if err := config.DB.Order("created_at desc").Find(&status).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render status menggunakan layout main.html
	c.HTML(http.StatusOK, "status", gin.H{
		"title":      "Status",
		"user":       employee, // seluruh row employee yang login (boleh nil)
		"status":     status,   // seluruh row employee yang login (boleh nil)
		"activePage": "status",
		"success":    success,
	})
}

// GET /status/create
func (dc *StatusController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "status_add", gin.H{
		"title":      "Add Status",
		"activePage": "status",
		"action":     "/status",
		"method":     "POST",
	})
}

// POST /status
func (dc *StatusController) Store(c *gin.Context) {
	var input models.Status

	// ambil field "status" dari form
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "status_add", gin.H{
			"title":      "Add Status",
			"activePage": "status",
			"error":      "Name status required",
			"data":       input,
			"action":     "/status",
			"method":     "POST",
		})
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Status success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/status")
}

// GET /status/:id/status
func (dc *StatusController) Edit(c *gin.Context) {
	id := c.Param("id")

	var status models.Status
	if err := config.DB.First(&status, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Status not found")
		return
	}

	c.HTML(http.StatusOK, "status_edit", gin.H{
		"title":      "Edit Status",
		"activePage": "status",
		"data":       status,
		"action":     "/status/" + id,
		"method":     "POST",
		"isEdit":     true,
	})
}

// POST /status/:id
func (dc *StatusController) Update(c *gin.Context) {
	id := c.Param("id")

	var status models.Status
	if err := config.DB.First(&status, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Status not found")
		return
	}

	var input struct {
		Status string `form:"status" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "status_edit", gin.H{
			"title":      "Edit Status",
			"activePage": "status",
			"error":      "Name status required",
			"data":       status,
			"action":     "/status/" + id,
			"method":     "POST",
			"isEdit":     true,
		})
		return
	}

	status.Status = input.Status

	if err := config.DB.Save(&status).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Status success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/status")
}

func (dc *StatusController) Delete(c *gin.Context) {
	id := c.Param("id")

	var status models.Status
	if err := config.DB.First(&status, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Status not found")
		return
	}

	if err := config.DB.Delete(&status).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Status success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/status")
}
