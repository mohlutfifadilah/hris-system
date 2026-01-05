package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeController struct{}

func NewEmployeeController() *EmployeeController {
	return &EmployeeController{}
}

// Parse UUID dan return pointer (nil jika kosong/error)
func parseUUIDPtr(s string) (*uuid.UUID, error) {
    if s == "" {
        return nil, nil
    }
    parsed, err := uuid.Parse(s)
    if err != nil {
        return nil, err
    }
    return &parsed, nil
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

// GET /employee/create
func (dc *EmployeeController) Create(c *gin.Context) {

	var bloods []models.Blood
	if err := config.DB.Order("created_at desc").Find(&bloods).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	var religions []models.Religion
	if err := config.DB.Order("created_at desc").Find(&religions).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	var identities []models.TypeIdentity
	if err := config.DB.Order("created_at desc").Find(&identities).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}
	var banks []models.TypeBank
	if err := config.DB.Order("created_at desc").Find(&banks).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	c.HTML(http.StatusOK, "employee_add", gin.H{
		"title":      "Add Employee",
		"activePage": "employee",
		"form":       map[string]string{},
		"errors":     map[string]string{},
		"bloods":     bloods,
		"religions":  religions,
		"identities": identities,
		"banks":      banks,
		"action":     "/employee",
		"method":     "POST",
	})
}

// POST /employee
func (dc *EmployeeController) Store(c *gin.Context) {
    form := map[string]string{
        "id_employee":      c.PostForm("id_employee"),
        "name":             c.PostForm("name"),
        "email":            c.PostForm("email"),
        "gender":           c.PostForm("gender"),
        "work_email":       c.PostForm("work_email"),
        "place_of_birth":   c.PostForm("place_of_birth"),
        "date_of_birth":    c.PostForm("date_of_birth"),
        "join_date":        c.PostForm("join_date"),
        "blood_type":       c.PostForm("blood_type"),
        "religion":         c.PostForm("religion"),
        "citizenship":      c.PostForm("citizenship"),
        "marial_status":    c.PostForm("marial_status"),
        "id_type_bank":     c.PostForm("id_type_bank"),
        "no_account":       c.PostForm("no_account"),
        "id_type_identity": c.PostForm("id_type_identity"),
        "no":               c.PostForm("no"),
        "evidence_link":    c.PostForm("evidence_link"),
    }

    errors := map[string]string{}

    // Validasi
    if strings.TrimSpace(form["name"]) == "" {
        errors["name"] = "Full Name wajib diisi"
    }
    if strings.TrimSpace(form["email"]) == "" {
        errors["email"] = "Email wajib diisi"
    }

    // Validasi photo
    file, fileErr := c.FormFile("photo")
    if fileErr == nil {
        ct := file.Header.Get("Content-Type")
        if !strings.HasPrefix(ct, "image/") {
            errors["photo"] = "Photo harus berupa gambar (image/*)"
        }
        if file.Size > 2*1024*1024 {
            errors["photo"] = "Photo maksimal 2MB"
        }
    }

    // Parse UUID (return pointer)
	bloodType, err := parseUUIDPtr(form["blood_type"])
	if err != nil {
		errors["blood_type"] = "Blood type UUID tidak valid"
	}

	idReligion, err := parseUUIDPtr(form["religion"])
	if err != nil {
		errors["religion"] = "Religion UUID tidak valid"
	}

	idTypeIdentity, err := parseUUIDPtr(form["id_type_identity"])
	if err != nil {
		errors["id_type_identity"] = "Identity type UUID tidak valid"
	}

	idTypeBank, err := parseUUIDPtr(form["id_type_bank"])
	if err != nil {
		errors["id_type_bank"] = "Bank type UUID tidak valid"
	}

    // Parse date
    var dateOfBirth time.Time
    if form["date_of_birth"] != "" {
        dateOfBirth, err = time.Parse("2006-01-02", form["date_of_birth"])
        if err != nil {
            errors["date_of_birth"] = "Format tanggal lahir tidak valid"
        }
    }

    // Parse MaritalStatus (boolean)
    maritalStatus := form["marial_status"] == "Married"

    // Kalau ada error, render ulang
    if len(errors) > 0 {
        var bloods []models.Blood
        var religions []models.Religion
        var banks []models.TypeBank
        var identities []models.TypeIdentity

        config.DB.Order("created_at desc").Find(&bloods)
        config.DB.Order("created_at desc").Find(&religions)
        config.DB.Order("created_at desc").Find(&banks)
        config.DB.Order("created_at desc").Find(&identities)

        c.HTML(http.StatusBadRequest, "employee_add", gin.H{
            "title":      "Add Employee",
            "action":     "/employee",
            "form":       form,
            "errors":     errors,
            "swalError":  "Ada input yang belum valid.",
            "bloods":     bloods,
            "religions":  religions,
            "banks":      banks,
            "identities": identities,
        })
        return
    }

    // Upload photo
    photoURL := ""
    if fileErr == nil {
        _ = os.MkdirAll("./static/assets/employee_photo", 0755)

        ext := strings.ToLower(filepath.Ext(file.Filename))
        if ext == "" {
            ext = ".jpg"
        }

        filename := fmt.Sprintf("emp_%d%s", time.Now().UnixNano(), ext)
        dst := filepath.Join("./static/assets/employee_photo", filename)

        if err := c.SaveUploadedFile(file, dst); err != nil {
            c.String(http.StatusInternalServerError, "upload failed: "+err.Error())
            return
        }

        photoURL = "/static/assets/employee_photo/" + filename
    }

    // Get company
    var company models.Company
    config.DB.First(&company)

    // Create employee
    emp := models.Employee{
        IDCompany:     &company.ID,
        IDStaffing:    nil,
        IDContact:     nil,
        IDIdentity:    idTypeIdentity,
        IDBankAccount: idTypeBank,
        IDBlood:       bloodType,
        IDReligion:    idReligion,
        WorkEmail:     form["work_email"],
        Email:         form["email"],
        Name:          form["name"],
        Photo:         photoURL,
        IDEmployee:    form["id_employee"],
        Gender:        form["gender"],
        Citizenship:   form["citizenship"],
        PlaceOfBirth:  form["place_of_birth"],
        DateOfBirth:   dateOfBirth,
        MarialStatus:  maritalStatus,
        JoinDate:      form["join_date"],
        EvidenceLink:  form["evidence_link"],
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    if err := config.DB.Create(&emp).Error; err != nil {
        c.String(http.StatusInternalServerError, "create employee failed: "+err.Error())
        return
    }

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

    c.Redirect(http.StatusSeeOther, "/employee")
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
