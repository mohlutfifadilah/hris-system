package controllers

import (
	"log"
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type EmployeeController struct{}

func NewEmployeeController() *EmployeeController {
	return &EmployeeController{}
}

// Index - Tampilkan halaman profile
func (dc *EmployeeController) Index(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	var employee models.Employee
	if err := config.DB.
		Where("id = ?", currentUser.ID).
		First(&employee).Error; err != nil {
		// handle error (404, dll)
		c.String(http.StatusInternalServerError, "employee not found")
		return
	}

	// 1) Ambil row company lengkap dari DB
	var company models.Company
	if err := config.DB.
		Where("id = ?", employee.IDCompany).
		First(&company).Error; err != nil {
		// handle error (404, dll)
		c.String(http.StatusInternalServerError, "company not found")
		return
	}

	// 2) Ambil row staffing lengkap dari DB
	var staffing models.Staffing
	if employee.IDStaffing != nil {
		if err := config.DB.
			Where("id = ?", employee.IDStaffing).
			First(&staffing).Error; err != nil {
			// handle error (404, dll)
			c.String(http.StatusInternalServerError, "staffing not found")
			return
		}
	} else {
		log.Println("Employee has no staffing (superadmin or unassigned)")
	}

	// 3) Ambil row contact lengkap dari DB
	var contact models.Contact
	if employee.IDContact != nil {
		if err := config.DB.
			Where("id = ?", employee.IDContact).
			First(&contact).Error; err != nil {
			// handle error (404, dll)
			c.String(http.StatusInternalServerError, "contact not found")
			return
		}
	} else {
		log.Println("Employee has no contact (superadmin or unassigned)")
	}

	var employees []models.Employee
	if err := config.DB.Order("created_at desc").Find(&employees).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render employee menggunakan layout main.html
	c.HTML(http.StatusOK, "employee", gin.H{
		"title":      "Employee",
		"user":       employee,  // seluruh row employee yang login (boleh nil)
		"users":      employees, // seluruh row employee yang login (boleh nil)
		"activePage": "employee",
		"success":    success,
	})
}

// GET /departments/create
func (dc *EmployeeController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "department_add", gin.H{
		"title":      "Add Department",
		"activePage": "department",
		"action":     "/departments",
		"method":     "POST",
	})
}

// POST /departments
func (dc *EmployeeController) Store(c *gin.Context) {
	var input models.Department

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
func (dc *EmployeeController) Edit(c *gin.Context) {
	id := c.Param("id")

	var dept models.Department
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
func (dc *EmployeeController) Update(c *gin.Context) {
	id := c.Param("id")

	var dept models.Department
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

func (dc *EmployeeController) Delete(c *gin.Context) {
	id := c.Param("id")

	var dept models.Department
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
