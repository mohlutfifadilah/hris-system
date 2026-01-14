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
	"hris-system/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

    // Buat map untuk simpan no_hp berdasarkan employee ID
    contactMap := make(map[uuid.UUID]string)

    for _, emp := range employees {
        if emp.IDContact != nil {
            var contact models.Contact
            if err := config.DB.First(&contact, "id = ?", emp.IDContact).Error; err == nil {
                contactMap[emp.ID] = contact.NoHp
            }
        }
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
		"contactMap": contactMap, // seluruh row employee yang login (boleh nil)
		"activePage": "employee",
		"success":    success,
	})
}

// GET /employee/create
func (dc *EmployeeController) Create(c *gin.Context) {
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
        "user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /employee
func (dc *EmployeeController) Store(c *gin.Context) {
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

    form := map[string]string{
        "id_employee":      c.PostForm("id_employee"),
        "name":             c.PostForm("name"),
        "email":            c.PostForm("email"),
        "gender":           c.PostForm("gender"),
        "citizenship":      c.PostForm("citizenship"),
        "marital_status":   c.PostForm("marital_status"),
        "blood_type":       c.PostForm("blood_type"),
        "religion":         c.PostForm("religion"),
        "work_email":       c.PostForm("work_email"),
        "evidence_link":    c.PostForm("evidence_link"),
        "place_of_birth":   c.PostForm("place_of_birth"),
        "date_of_birth":    c.PostForm("date_of_birth"),
        "join_date":        c.PostForm("join_date"),
        "id_type_bank":     c.PostForm("id_type_bank"),
        "no_account":       c.PostForm("no_account"),
        "no_kpj_bpjs":       c.PostForm("no_kpj_bpjs"),
        "no_bpjs_kes":       c.PostForm("no_bpjs_kes"),
        "no_bpjs_jk":       c.PostForm("no_bpjs_jk"),
        "no_npwp_limabelas":       c.PostForm("no_npwp_limabelas"),
        "no_npwp_enambelas":       c.PostForm("no_npwp_enambelas"),
        "ptkp":       c.PostForm("ptkp"),
        "id_type_identity": c.PostForm("id_type_identity"),
        "no":       c.PostForm("no"),
        "address_identity":       c.PostForm("address_identity"),
        "address_domicile":       c.PostForm("address_domicile"),
        "no_hp":       c.PostForm("no_hp"),
        "no_emergency_contact":       c.PostForm("no_emergency_contact"),
        "emergency_contact_name":       c.PostForm("emergency_contact_name"),
        "emergency_relation":       c.PostForm("emergency_relation"),
    }

    errors := map[string]string{}

    // Validasi photo
    file, fileErr := c.FormFile("photo")
    if fileErr == nil {
        ct := file.Header.Get("Content-Type")
        if !strings.HasPrefix(ct, "image/") {
            errors["photo"] = "Photo must image (image/*)"
        }
        if file.Size > 2*1024*1024 {
            errors["photo"] = "Max Size 2MB"
        }
    }

    // Parse UUID (return pointer)
	bloodType, err := parseUUIDPtr(form["blood_type"])
	if err != nil {
		errors["blood_type"] = "Blood type UUID not valid"
	}

	idReligion, err := parseUUIDPtr(form["religion"])
	if err != nil {
		errors["religion"] = "Religion UUID not valid"
	}

	idTypeIdentity, err := parseUUIDPtr(form["id_type_identity"])
	if err != nil {
		errors["id_type_identity"] = "Identity type UUID not valid"
	}

	idTypeBank, err := parseUUIDPtr(form["id_type_bank"])
	if err != nil {
		errors["id_type_bank"] = "Bank type UUID not valid"
	}

    // Parse date of birth
    var dateOfBirth time.Time
    if form["date_of_birth"] != "" {
        dateOfBirth, err = time.Parse("2006-01-02", form["date_of_birth"])
        if err != nil {
            errors["date_of_birth"] = "Date of birth format not valid"
        }
    }

    // Parse join date
    var joinDate time.Time
    if form["join_date"] != "" {
        joinDate, err = time.Parse("2006-01-02", form["join_date"])
        if err != nil {
            errors["join_date"] = "Join date format not valid"
        }
    }

    var maritalStatus bool
	maritalStatusStr := form["marital_status"]

	if maritalStatusStr == "" {
		errors["marital_status"] = "Marital status is required"
	} else if maritalStatusStr == "Married" {
		maritalStatus = true
	} else if maritalStatusStr == "Not Married" {
		maritalStatus = false
	} else {
		errors["marital_status"] = "Invalid marital status value"
	}

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
            "swalError":  "There are some invalid inputs.",
            "bloods":     bloods,
            "religions":  religions,
            "banks":      banks,
            "identities": identities,
            "user":       employee, // seluruh row employee yang login (boleh nil)
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

	// ========== START TRANSACTION ==========
    tx := config.DB.Begin()

	// 1. Create Staffing
    staffingID := uuid.New()
    staffing := models.Staffing{
        ID:              staffingID,
        NoBpjsKes:       form["no_bpjs_kes"],
        NoBpjsJk:        form["no_bpjs_jk"],
        NoKpjBpjs:       form["no_kpj_bpjs"],
        NoNpwpLimabelas: form["no_npwp_limabelas"],
        NoNpwpEnambelas: form["no_npwp_enambelas"],
        Ptkp:            form["ptkp"],
    }

    if err := tx.Create(&staffing).Error; err != nil {
        tx.Rollback()
        log.Println("Error creating staffing:", err)
        c.String(http.StatusInternalServerError, "employee_add", gin.H{
            "title":  "Add Employee",
            "errors": map[string]string{"form": "Failed to create staffing data"},
            "form":   form,
        })
        return
    }

	// 2. Create Address
    addressID := uuid.New()
    address := models.Address{
        ID:                   addressID,
        AddressIdentity:      form["address_identity"],
        AddressDomicile:      form["address_domicile"],
    }

    if err := tx.Create(&address).Error; err != nil {
        tx.Rollback()
        log.Println("Error creating address:", err)
        c.HTML(http.StatusInternalServerError, "employee_add", gin.H{
            "title":  "Add Employee",
            "errors": map[string]string{"form": "Failed to create address data"},
            "form":   form,
        })
        return
    } 

	// 3. Create Contact
    contactID := uuid.New()
    contact := models.Contact{
        ID:                   contactID,
        IDAddress:            &addressID,
        NoHp:                 form["no_hp"],
        EmergencyContactName: form["emergency_contact_name"],
        NoEmergencyContact:   form["no_emergency_contact"],
        EmergencyRelation:    form["emergency_relation"],
    }

    if err := tx.Create(&contact).Error; err != nil {
        tx.Rollback()
        log.Println("Error creating contact:", err)
        c.HTML(http.StatusInternalServerError, "employee_add", gin.H{
            "title":  "Add Employee",
            "errors": map[string]string{"form": "Failed to create contact data"},
            "form":   form,
        })
        return
    }

    // Get company
    var company models.Company
    tx.First(&company)

	// Generate random password
    plainPassword := utils.GenerateRandomPassword()
    
    // Hash password untuk disimpan di database
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    if err != nil {
        log.Println("Error hashing password:", err)
        c.HTML(http.StatusInternalServerError, "employee_add", gin.H{
            "title":  "Add Employee",
            "errors": map[string]string{"form": "Failed to generate password"},
            "form":   form,
        })
        return
    }

    // Create employee
    emp := models.Employee{
        IDCompany:     &company.ID,
        IDStaffing:    &staffingID,
        IDContact:     &contactID,
        IDIdentity:    idTypeIdentity,
        IDBankAccount: idTypeBank,
        IDBlood:       bloodType,
        IDReligion:    idReligion,
        WorkEmail:     form["work_email"],
        Email:         form["email"],
		Password:      string(hashedPassword), // kosongkan dulu, nanti bisa direset
        Name:          form["name"],
        Photo:         photoURL,
		IDEmployee:    form["id_employee"],
        Gender:        form["gender"],
        Citizenship:   form["citizenship"],
        PlaceOfBirth:  form["place_of_birth"],
        DateOfBirth:   dateOfBirth,
        MaritalStatus: maritalStatus,
        JoinDate:      joinDate,
        EvidenceLink:  form["evidence_link"],
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    if err := tx.Create(&emp).Error; err != nil {
        c.String(http.StatusInternalServerError, "create employee failed: "+err.Error())
        return
    }

	// ========== COMMIT TRANSACTION ==========
    if err := tx.Commit().Error; err != nil {
        log.Println("Error committing transaction:", err)
        c.HTML(http.StatusInternalServerError, "employee_add", gin.H{
            "title":  "Add Employee",
            "errors": map[string]string{"form": "Failed to save data"},
            "form":   form,
        })
        return
    }

    // ========== SEND PASSWORD EMAIL ==========
    emailConfig := utils.EmailConfig{
        SMTPHost:     "smtp.gmail.com",           // Ganti dengan SMTP server kamu
        SMTPPort:     587,                        // Port SMTP (587 untuk TLS, 465 untuk SSL)
        SMTPUsername: "mohlutfifadilah23@gmail.com",  // Email admin
        SMTPPassword: "dklf kykb girq cymb",      // Password email admin (atau App Password)
        FromEmail:    "mohlutfifadilah23@gmail.com",
        FromName:     "HRIS Admin",
    }
	

    // Send email (async, tidak block response)
    go func() {
        if err := utils.SendPasswordEmail(form["work_email"], form["name"], plainPassword, emailConfig); err != nil {
            log.Printf("Failed to send password email: %v", err)
        }
    }()

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

    c.Redirect(http.StatusFound, "/employee")
}

// GET /departments/:id/edit
func (dc *EmployeeController) Edit(c *gin.Context) {
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
        "user":       employee, // seluruh row employee yang login (boleh nil)
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
