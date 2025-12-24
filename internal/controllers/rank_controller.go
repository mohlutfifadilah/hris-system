package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
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

	var departments []models.DepartmentHistory
	if err := config.DB.Order("created_at desc").Find(&departments).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render department menggunakan layout main.html
	c.HTML(http.StatusOK, "department", gin.H{
		"title":       "Department",
		"user":        employee,    // seluruh row employee yang login (boleh nil)
		"departments": departments, // seluruh row employee yang login (boleh nil)
		"activePage":  "department",
		"success":     success,
	})
}

// GET /departments/create
func (dc *DepartmentController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "department_add", gin.H{
		"title":      "Add Department",
		"activePage": "department",
		"action":     "/departments",
		"method":     "POST",
	})
}

// POST /departments
func (dc *DepartmentController) Store(c *gin.Context) {
	var input models.DepartmentHistory

	// ambil field "department" dari form
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "department_add", gin.H{
			"title":      "Add Department",
			"activePage": "department",
			"error":      "Name department required",
			"data":       input,
			"action":     "/departments",
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
	session.Set("flash_success", "Department success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/department")
}

// GET /departments/:id/edit
func (dc *DepartmentController) Edit(c *gin.Context) {
	id := c.Param("id")

	var dept models.DepartmentHistory
	if err := config.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Department not found")
		return
	}

	c.HTML(http.StatusOK, "department_edit", gin.H{
		"title":      "Edit Department",
		"activePage": "department",
		"data":       dept,
		"action":     "/departments/" + id,
		"method":     "POST",
		"isEdit":     true,
	})
}

// POST /departments/:id
func (dc *DepartmentController) Update(c *gin.Context) {
	id := c.Param("id")

	var dept models.DepartmentHistory
	if err := config.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Department not found")
		return
	}

	var input struct {
		Department string `form:"department" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "department_edit", gin.H{
			"title":      "Edit Department",
			"activePage": "department",
			"error":      "Name department required",
			"data":       dept,
			"action":     "/departments/" + id,
			"method":     "POST",
			"isEdit":     true,
		})
		return
	}

	dept.Department = input.Department

	if err := config.DB.Save(&dept).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Department success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/department")
}

func (dc *DepartmentController) Delete(c *gin.Context) {
	id := c.Param("id")

	var dept models.DepartmentHistory
	if err := config.DB.First(&dept, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Department not found")
		return
	}

	if err := config.DB.Delete(&dept).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Department success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/department")
}
