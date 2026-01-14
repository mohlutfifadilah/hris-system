package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type GradingController struct{}

func NewGradingController() *GradingController {
	return &GradingController{}
}

// Index - Tampilkan halaman
func (dc *GradingController) Index(c *gin.Context) {
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

	var grading []models.Grading
	if err := config.DB.Order("created_at desc").Find(&grading).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render grading menggunakan layout main.html
	c.HTML(http.StatusOK, "grading", gin.H{
		"title":      "Grading",
		"user":       employee, // seluruh row employee yang login (boleh nil)
		"grading":    grading,  // seluruh row employee yang login (boleh nil)
		"activePage": "grading",
		"success":    success,
	})
}

// GET /grading/create
func (dc *GradingController) Create(c *gin.Context) {
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

	c.HTML(http.StatusOK, "grading_add", gin.H{
		"title":      "Add Grading",
		"activePage": "grading",
		"action":     "/grading",
		"method":     "POST",
		"user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /grading
func (dc *GradingController) Store(c *gin.Context) {
	var input models.Grading

	// ambil field "grading" dari form
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "grading_add", gin.H{
			"title":      "Add Grading",
			"activePage": "grading",
			"error":      "Name grading required",
			"data":       input,
			"action":     "/grading",
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
	session.Set("flash_success", "Grading success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/grading")
}

// GET /grading/:id/grading
func (dc *GradingController) Edit(c *gin.Context) {
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

	id := c.Param("id")

	var grading models.Grading
	if err := config.DB.First(&grading, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Grading not found")
		return
	}

	c.HTML(http.StatusOK, "grading_edit", gin.H{
		"title":      "Edit Grading",
		"activePage": "grading",
		"data":       grading,
		"action":     "/grading/" + id,
		"method":     "POST",
		"isEdit":     true,
		"user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /grading/:id
func (dc *GradingController) Update(c *gin.Context) {
	id := c.Param("id")

	var grading models.Grading
	if err := config.DB.First(&grading, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Grading not found")
		return
	}

	var input struct {
		Grading string `form:"grading" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "grading_edit", gin.H{
			"title":      "Edit Grading",
			"activePage": "grading",
			"error":      "Name grading required",
			"data":       grading,
			"action":     "/grading/" + id,
			"method":     "POST",
			"isEdit":     true,
		})
		return
	}

	grading.Grading = input.Grading

	if err := config.DB.Save(&grading).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Grading success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/grading")
}

func (dc *GradingController) Delete(c *gin.Context) {
	id := c.Param("id")

	var grading models.Grading
	if err := config.DB.First(&grading, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Grading not found")
		return
	}

	if err := config.DB.Delete(&grading).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Grading success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/grading")
}
