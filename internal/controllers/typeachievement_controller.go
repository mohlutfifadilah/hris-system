package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type TypeAchievement struct{}

func NewTypeAchievement() *TypeAchievement {
	return &TypeAchievement{}
}

// Index - Tampilkan halaman
func (dc *TypeAchievement) Index(c *gin.Context) {
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

	var typeAchievement []models.TypeAchievement
	if err := config.DB.Order("created_at desc").Find(&typeAchievement).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render typeAchievement menggunakan layout main.html
	c.HTML(http.StatusOK, "typeAchievement", gin.H{
		"title":           "Type Achievement",
		"user":            employee,        // seluruh row employee yang login (boleh nil)
		"typeAchievement": typeAchievement, // seluruh row employee yang login (boleh nil)
		"activePage":      "typeAchievement",
		"success":         success,
	})
}

// GET /typeAchievement/create
func (dc *TypeAchievement) Create(c *gin.Context) {

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

	c.HTML(http.StatusOK, "typeAchievement_add", gin.H{
		"title":      "Add Type Achievement",
		"activePage": "typeAchievement",
		"action":     "/typeAchievement",
		"method":     "POST",
		"user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /typeAchievement
func (dc *TypeAchievement) Store(c *gin.Context) {
	var input models.TypeAchievement

	// ambil field "type_achievement" dari form
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "typeAchievement_add", gin.H{
			"title":      "Add Type Achievement",
			"activePage": "typeAchievement",
			"error":      "Name type achievement required",
			"data":       input,
			"action":     "/typeAchievement",
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
	session.Set("flash_success", "Type Achievement success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/typeAchievement")
}

// GET /typeAchievement/:id/status
func (dc *TypeAchievement) Edit(c *gin.Context) {
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

	var typeAchievement models.TypeAchievement
	if err := config.DB.First(&typeAchievement, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Type Achievement not found")
		return
	}

	c.HTML(http.StatusOK, "typeAchievement_edit", gin.H{
		"title":      "Edit Type Achievement",
		"activePage": "typeAchievement",
		"data":       typeAchievement,
		"action":     "/typeAchievement/" + id,
		"method":     "POST",
		"isEdit":     true,
		"user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /typeAchievement/:id
func (dc *TypeAchievement) Update(c *gin.Context) {
	id := c.Param("id")

	var typeAchievement models.TypeAchievement
	if err := config.DB.First(&typeAchievement, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Type Achievement not found")
		return
	}

	var input struct {
		Type string `form:"type" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "typeAchievement_edit", gin.H{
			"title":      "Edit Type Achievement",
			"activePage": "typeAchievement",
			"error":      "Name type achievement required",
			"data":       typeAchievement,
			"action":     "/typeAchievement/" + id,
			"method":     "POST",
			"isEdit":     true,
		})
		return
	}

	typeAchievement.Type = input.Type

	if err := config.DB.Save(&typeAchievement).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Type Achievement success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/typeAchievement")
}

func (dc *TypeAchievement) Delete(c *gin.Context) {
	id := c.Param("id")

	var typeAchievement models.TypeAchievement
	if err := config.DB.First(&typeAchievement, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Type Achievement not found")
		return
	}

	if err := config.DB.Delete(&typeAchievement).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Type Achievement success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/typeAchievement")
}
