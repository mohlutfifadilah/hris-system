package controllers

import (
	"log"
	"net/http"
	"time"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AchievementController struct{}

func NewAchievementController() *AchievementController {
	return &AchievementController{}
}

// Index - Tampilkan halaman profile
func (dc *AchievementController) Index(c *gin.Context) {
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

	var achievement []models.Achievement
	if err := config.DB.Order("created_at desc").Find(&achievement).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

    // Buat map untuk simpan employee berdasarkan employee ID
    employeeMap := make(map[uuid.UUID]string)

    for _, emp := range achievement {
        if emp.IDEmployee != nil {
            var employee models.Employee
            if err := config.DB.First(&employee, "id = ?", emp.IDEmployee).Error; err == nil {
                employeeMap[emp.ID] = employee.ID
            }
        }
    }

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render achievement menggunakan layout main.html
	c.HTML(http.StatusOK, "achievement", gin.H{
		"title":      "Achievement",
		"user":       employee,  // seluruh row employee yang login (boleh nil)
		"achievement":      achievement, // seluruh row employee yang login (boleh nil)
		"contactMap": contactMap, // seluruh row employee yang login (boleh nil)
		"activePage": "achievement",
		"success":    success,
	})
}

// GET /achievement/create
func (dc *AchievementController) Create(c *gin.Context) {

	var employees []models.Employee
	if err := config.DB.Order("created_at desc").Find(&employees).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	var typeAchievement []models.TypeAchievement
	if err := config.DB.Order("created_at desc").Find(&typeAchievement).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	c.HTML(http.StatusOK, "achievement_add", gin.H{
		"title":      "Add Achievement",
		"activePage": "achievement",
		"form":       map[string]string{},
		"errors":     map[string]string{},
		"employees":     employees,
		"typeAchievement":     typeAchievement,
		"action":     "/achievement",
		"method":     "POST",
	})
}

// POST /achievement
func (dc *AchievementController) Store(c *gin.Context) {
    form := map[string]string{
        "employee":      c.PostForm("employee"),
        "type_achievement":      c.PostForm("type_achievement"),
        "date":      c.PostForm("date"),
        "title":      c.PostForm("title"),
        "description":      c.PostForm("description"),
        "evidence_link":      c.PostForm("evidence_link"),
    }

    errors := map[string]string{}

    // Parse UUID (return pointer)
	employee, err := parseUUIDPtr(form["employee"])
	if err != nil {
		errors["employee"] = "Employee UUID not valid"
	}

	type_achievement, err := parseUUIDPtr(form["type_achievement"])
	if err != nil {
		errors["type_achievement"] = "Type Achievement UUID not valid"
	}

    // Parse date
    var date time.Time
    if form["date"] != "" {
        date, err = time.Parse("2006-01-02", form["date"])
        if err != nil {
            errors["date"] = "Date format not valid"
        }
    }

    // Kalau ada error, render ulang
    if len(errors) > 0 {
        var employees []models.Employee
        var typeAchievement []models.TypeAchievement

        config.DB.Order("created_at desc").Find(&employees)
        config.DB.Order("created_at desc").Find(&typeAchievement)

        c.HTML(http.StatusBadRequest, "achievement_add", gin.H{
            "title":      "Add Achievement",
            "action":     "/achievement",
            "form":       form,
            "errors":     errors,
            "swalError":  "There are some invalid inputs.",
            "employees":     employees,
            "typeAchievement":     typeAchievement,
            "method":     "POST",})
        return
    }

	// ========== START TRANSACTION ==========
    tx := config.DB.Begin()

    // Create achievement
    emp := models.Achievement{
        IDEmployee:     employee,
        IDTypeAchievement:    type_achievement,
        Date: date,
        Title:  form["title"],
        Description:  form["description"],
        EvidenceLink:  form["evidence_link"],
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    if err := tx.Create(&emp).Error; err != nil {
        c.String(http.StatusInternalServerError, "create achievement failed: "+err.Error())
        return
    }

	// ========== COMMIT TRANSACTION ==========
    if err := tx.Commit().Error; err != nil {
        log.Println("Error committing transaction:", err)
        c.HTML(http.StatusInternalServerError, "achievement_add", gin.H{
            "title":  "Add Achievement",
            "errors": map[string]string{"form": "Failed to save data"},
            "form":   form,
        })
        return
    }

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

    c.Redirect(http.StatusSeeOther, "/achievement")
}

// GET /departments/:id/edit
func (dc *AchievementController) Edit(c *gin.Context) {
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
func (dc *AchievementController) Update(c *gin.Context) {
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

func (dc *AchievementController) Delete(c *gin.Context) {
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
