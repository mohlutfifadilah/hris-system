package controllers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
	if err := config.DB.Where("id != ?", currentUser.ID).Order("created_at desc").Find(&employees).Error; err != nil {
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

	// CareerMap - INI YANG PENTING
	type CareerInfo struct {
		Position   string
		Department string
	}
	careerMap := make(map[string]CareerInfo)

	for _, emp := range employees {
		var result struct {
			Position   string
			Department string
		}

		query := fmt.Sprintf(`
            SELECT c.position, d.department
            FROM career_history c
            LEFT JOIN department d ON d.id::text = c.id_department
            WHERE c.id_employee = '%s'
            ORDER BY c.effective_date DESC
            LIMIT 1
        `, emp.ID.String())

		if err := config.DB.Raw(query).Scan(&result).Error; err == nil {
			careerMap[emp.ID.String()] = CareerInfo{
				Position:   result.Position,
				Department: result.Department,
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
		"user":       employee,   // seluruh row employee yang login (boleh nil)
		"users":      employees,  // seluruh row employee yang login (boleh nil)
		"contactMap": contactMap, // seluruh row employee yang login (boleh nil)
		"careerMap":  careerMap,  // seluruh row employee yang login (boleh nil)
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
		"id_employee":            c.PostForm("id_employee"),
		"name":                   c.PostForm("name"),
		"email":                  c.PostForm("email"),
		"gender":                 c.PostForm("gender"),
		"citizenship":            c.PostForm("citizenship"),
		"marital_status":         c.PostForm("marital_status"),
		"blood_type":             c.PostForm("blood_type"),
		"religion":               c.PostForm("religion"),
		"work_email":             c.PostForm("work_email"),
		"place_of_birth":         c.PostForm("place_of_birth"),
		"date_of_birth":          c.PostForm("date_of_birth"),
		"join_date":              c.PostForm("join_date"),
		"id_type_bank":           c.PostForm("id_type_bank"),
		"no_account":             c.PostForm("no_account"),
		"no_kpj_bpjs":            c.PostForm("no_kpj_bpjs"),
		"no_bpjs_kes":            c.PostForm("no_bpjs_kes"),
		"no_bpjs_jk":             c.PostForm("no_bpjs_jk"),
		"no_npwp_limabelas":      c.PostForm("no_npwp_limabelas"),
		"no_npwp_enambelas":      c.PostForm("no_npwp_enambelas"),
		"ptkp":                   c.PostForm("ptkp"),
		"id_type_identity":       c.PostForm("id_type_identity"),
		"no":                     c.PostForm("no"),
		"address_identity":       c.PostForm("address_identity"),
		"address_domicile":       c.PostForm("address_domicile"),
		"no_hp":                  c.PostForm("no_hp"),
		"no_emergency_contact":   c.PostForm("no_emergency_contact"),
		"emergency_contact_name": c.PostForm("emergency_contact_name"),
		"emergency_relation":     c.PostForm("emergency_relation"),
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
		ID:              addressID,
		AddressIdentity: form["address_identity"],
		AddressDomicile: form["address_domicile"],
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

	// Create Identity
	identityID := uuid.New()
	identity := models.Identity{
		ID:             identityID,
		IDTypeIdentity: idTypeIdentity,
		No:             form["no"],
	}

	if err := tx.Create(&identity).Error; err != nil {
		tx.Rollback()
		log.Println("Error creating identity:", err)
		c.String(http.StatusInternalServerError, "employee_add", gin.H{
			"title":  "Add Employee",
			"errors": map[string]string{"form": "Failed to create identity data"},
			"form":   form,
		})
		return
	}

	// Create Bank Account
	bankAccountID := uuid.New()
	bankAccount := models.BankAccount{
		ID:         bankAccountID,
		IDTypeBank: idTypeBank,
		NoAccount:  form["no_account"],
	}

	if err := tx.Create(&bankAccount).Error; err != nil {
		tx.Rollback()
		log.Println("Error creating bank account:", err)
		c.String(http.StatusInternalServerError, "employee_add", gin.H{
			"title":  "Add Employee",
			"errors": map[string]string{"form": "Failed to create bank account data"},
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
		IDIdentity:    &identityID,
		IDBankAccount: &bankAccountID,
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
		IsAdmin:       false,
		JoinDate:      joinDate,
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
		SMTPHost:     "smtp.gmail.com",              // Ganti dengan SMTP server kamu
		SMTPPort:     587,                           // Port SMTP (587 untuk TLS, 465 untuk SSL)
		SMTPUsername: "mohlutfifadilah23@gmail.com", // Email admin
		SMTPPassword: "dklf kykb girq cymb",         // Password email admin (atau App Password)
		FromEmail:    "mohlutfifadilah23@gmail.com",
		FromName:     "HRIS Admin",
	}

	// Send email (async, tidak block response)
	go func() {
		if err := utils.SendPasswordEmail(form["work_email"], form["name"], plainPassword, emailConfig); err != nil {
			log.Printf("Failed to send password email: %v", err)
		}
	}()

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Employee success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/employee")
}

// GET /employee/:id
func (dc *EmployeeController) Show(c *gin.Context) {
	id := c.Param("id")

	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	// 1. Ambil employee
	var employee models.Employee
	if err := config.DB.Where("id = ?", id).First(&employee).Error; err != nil {
		c.String(http.StatusNotFound, "Employee not found")
		return
	}

	// 3. Ambil master yang direferensikan employee (company, blood, religion, bank, identity, contact, staffing)
	var company models.Company
	var blood models.Blood
	var religion models.Religion
	var bankAccount models.BankAccount
	var typeBank models.TypeBank
	var identity models.Identity
	var typeIdentity models.TypeIdentity
	var contact models.Contact
	var address models.Address
	var staffing models.Staffing

	// company
	if employee.IDCompany != nil {
		_ = config.DB.First(&company, "id = ?", employee.IDCompany).Error
	}
	// blood
	if employee.IDBlood != nil {
		_ = config.DB.First(&blood, "id = ?", employee.IDBlood).Error
	}
	// religion
	if employee.IDReligion != nil {
		_ = config.DB.First(&religion, "id = ?", employee.IDReligion).Error
	}
	// bank account + type_bank
	if employee.IDBankAccount != nil {
		if err := config.DB.First(&bankAccount, "id = ?", employee.IDBankAccount).Error; err == nil {
			if bankAccount.IDTypeBank != nil {
				_ = config.DB.First(&typeBank, "id = ?", bankAccount.IDTypeBank).Error
			}
		}
	}
	// identity + type_identity
	if employee.IDIdentity != nil {
		if err := config.DB.First(&identity, "id = ?", employee.IDIdentity).Error; err == nil {
			if identity.IDTypeIdentity != nil {
				_ = config.DB.First(&typeIdentity, "id = ?", identity.IDTypeIdentity).Error
			}
		}
	}
	// contact + address
	if employee.IDContact != nil {
		if err := config.DB.First(&contact, "id = ?", employee.IDContact).Error; err == nil {
			if contact.IDAddress != nil {
				_ = config.DB.First(&address, "id = ?", contact.IDAddress).Error
			}
		}
	}
	// staffing
	if employee.IDStaffing != nil {
		_ = config.DB.First(&staffing, "id = ?", employee.IDStaffing).Error
	}

	// 4. Ambil data turunan by id_employee: education, career, achievement, family
	var educations []models.Education
	var careers []models.Career
	var achievements []models.Achievement
	var families []models.Family

	_ = config.DB.Where("id_employee = ?", employee.ID).Find(&educations).Error
	_ = config.DB.Where("id_employee = ?", employee.ID).Find(&careers).Error
	_ = config.DB.Where("id_employee = ?", employee.ID).Find(&achievements).Error
	_ = config.DB.Where("id_employee = ?", employee.ID).Find(&families).Error

	// 5. (Opsional) pre-load master untuk career (status, grading, department) & family (religion)
	// kalau butuh label di view
	var statuses []models.Status
	var gradings []models.Grading
	var departments []models.Department
	var familyReligions []models.Religion

	if len(careers) > 0 {
		_ = config.DB.Find(&statuses).Error
		_ = config.DB.Find(&gradings).Error
		_ = config.DB.Find(&departments).Error
	}
	if len(families) > 0 {
		_ = config.DB.Find(&familyReligions).Error
	}

	// Ambil career dengan effective_date terbaru
	var career models.Career
	err := config.DB.Where("id_employee = ?", employee.ID).
		Order("effective_date desc").
		First(&career).Error // Menggunakan First untuk mendapatkan hanya satu record dengan effective_date terbaru
	if err != nil {
		log.Println("Error fetching career:", err)
	}

	// department untuk current career
	var odepartment models.Department
	if career.IDDepartment != nil {
		_ = config.DB.First(&odepartment, "id = ?", career.IDDepartment).Error
	}

	// grading untuk current career
	var ograding models.Grading
	if career.IDGrading != nil {
		_ = config.DB.First(&ograding, "id = ?", career.IDGrading).Error
	}

	// 1. Ambil SEMUA career employee (NO YEAR FILTER dulu)
	var allCareer []models.Career
	config.DB.Where("id_employee = ?", id).
		Order("effective_date ASC").
		Find(&allCareer)

	// 2. Cache untuk relasi names
	statusNames := make(map[string]string)
	gradingNames := make(map[string]string)
	departmentNames := make(map[string]string)

	for _, career := range allCareer {
		// Get status name
		if career.IDStatus != nil {
			statusKey := career.IDStatus.String()
			if _, exists := statusNames[statusKey]; !exists {
				var status models.Status
				if err := config.DB.Where("id = ?", career.IDStatus).First(&status).Error; err == nil {
					statusNames[statusKey] = status.Status
				}
			}
		}

		// Get grading name
		if career.IDGrading != nil {
			gradingKey := career.IDGrading.String()
			if _, exists := gradingNames[gradingKey]; !exists {
				var grading models.Grading
				if err := config.DB.Where("id = ?", career.IDGrading).First(&grading).Error; err == nil {
					gradingNames[gradingKey] = grading.Grading
				}
			}
		}

		// Get department name
		if career.IDDepartment != nil {
			deptKey := career.IDDepartment.String()
			if _, exists := departmentNames[deptKey]; !exists {
				var dept models.Department
				if err := config.DB.Where("id = ?", career.IDDepartment).First(&dept).Error; err == nil {
					departmentNames[deptKey] = dept.Department
				}
			}
		}
	}

	// 3. Group by YEAR only
	yearMap := make(map[int][]models.Career)
	for _, career := range allCareer {
		year := career.EffectiveDate.Year()
		yearMap[year] = append(yearMap[year], career)
	}

	// 4. Convert to sorted yearGroups
	var yearGroups []CareerYearGroup
	for year, careers := range yearMap {
		// Sort careers by date DESC within year
		sort.Slice(careers, func(i, j int) bool {
			return careers[i].EffectiveDate.Before(careers[j].EffectiveDate)
		})

		yearGroups = append(yearGroups, CareerYearGroup{
			Year:    year,
			Careers: careers,
		})
	}

	// Sort by year descending
	sort.Slice(yearGroups, func(i, j int) bool {
		return yearGroups[i].Year < yearGroups[j].Year
	})

	var status []models.Status
	if err := config.DB.Order("created_at desc").Find(&status).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	var grading []models.Grading
	if err := config.DB.Order("created_at desc").Find(&grading).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	var department []models.Department
	if err := config.DB.Order("created_at desc").Find(&department).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	// 2. Ambil SEMUA achievements employee (NO YEAR FILTER dulu)
	var allAchievements []models.Achievement
	config.DB.Where("id_employee = ?", id).
		Order("date ASC").
		Find(&allAchievements)

	// 2. Group by YEAR dari data yang ada
	yearMap2 := make(map[int][]AchievementByType)

	for _, ach := range allAchievements {
		year := ach.Date.Year()

		// Get or create type group for this year
		typeMap := yearMap2[year]

		// Cari apakah type sudah ada di year ini
		found := false
		for i := range typeMap {
			if typeMap[i].TypeID == *ach.IDTypeAchievement {
				typeMap[i].Achievements = append(typeMap[i].Achievements, ach)
				found = true
				break
			}
		}

		if !found {
			// Type baru, fetch type name
			var typeAch models.TypeAchievement
			config.DB.First(&typeAch, *ach.IDTypeAchievement)

			yearMap2[year] = append(typeMap, AchievementByType{
				TypeName:     typeAch.Type,
				TypeID:       *ach.IDTypeAchievement,
				Achievements: []models.Achievement{ach},
			})
		}
	}

	// 3. Convert to sorted yearGroups
	var yearGroups2 []YearGroup
	for year := range yearMap2 {
		yearGroups2 = append(yearGroups2, YearGroup{
			Year:       year,
			TypeGroups: yearMap2[year],
		})
	}

	// Sort by year
	sort.Slice(yearGroups2, func(i, j int) bool {
		return yearGroups2[i].Year < yearGroups2[j].Year
	})

	var types []models.TypeAchievement
	config.DB.Order("type ASC").Find(&types)

	// 5. Render template show/detail
	c.HTML(http.StatusOK, "employee_info", gin.H{
		"title":      "Employee",
		"activePage": "employee",

		"users":        employee,
		"company":      company,
		"blood":        blood,
		"religion":     religion,
		"bankAccount":  bankAccount,
		"typeBank":     typeBank,
		"identity":     identity,
		"typeIdentity": typeIdentity,
		"contact":      contact,
		"address":      address,
		"staffing":     staffing,
		"career":       career,
		"odepartment":  odepartment,
		"ograding":     ograding,
		"department":   department,
		"grading":      grading,

		"yearGroups":      yearGroups,
		"allCareer":       allCareer, // Backup untuk timeline dots
		"statusNames":     statusNames,
		"gradingNames":    gradingNames,
		"departmentNames": departmentNames,

		"yearGroupss":     yearGroups2,
		"types":           types,
		"allAchievements": allAchievements, // Backup untuk timeline dots

		"educations":   educations,
		"careers":      careers,
		"achievements": achievements,
		"families":     families,

		// master untuk mapping ID -> nama (dipakai di range di template)
		"statuses":        statuses,
		"gradings":        gradings,
		"departments":     departments,
		"familyReligions": familyReligions,
		"user":            currentUser,
	})
}

// GET /departments/:id/edit
func (dc *EmployeeController) Edit(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	if currentUser == nil {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	// 1) Ambil row employee lengkap dari DB
	var user models.Employee
	if err := config.DB.Where("id = ?", currentUser.ID).First(&user).Error; err != nil {
		c.String(http.StatusInternalServerError, "employee not found")
		return
	}

	id := c.Param("id")

	// Ambil employee yang akan di-edit
	var employee models.Employee
	if err := config.DB.First(&employee, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Employee not found")
		return
	}

	// Ambil relasi staffing
	var staffing models.Staffing
	if employee.IDStaffing != nil {
		config.DB.First(&staffing, "id = ?", employee.IDStaffing)
	}

	// Ambil relasi identity
	var identity models.Identity
	if employee.IDIdentity != nil {
		config.DB.First(&identity, "id = ?", employee.IDIdentity)
	}

	// Ambil relasi contact
	var contact models.Contact
	if employee.IDContact != nil {
		config.DB.First(&contact, "id = ?", employee.IDContact)
	}

	// Ambil relasi address dari contact
	var address models.Address
	if contact.IDAddress != nil {
		config.DB.First(&address, "id = ?", contact.IDAddress)
	}

	// Ambil relasi bank account
	var bankAccount models.BankAccount
	if employee.IDBankAccount != nil {
		config.DB.First(&bankAccount, "id = ?", employee.IDBankAccount)
	}

	// Ambil data dropdown
	var bloods []models.Blood
	config.DB.Find(&bloods)

	var religions []models.Religion
	config.DB.Find(&religions)

	var banks []models.TypeBank
	config.DB.Find(&banks)

	var identities []models.TypeIdentity
	config.DB.Find(&identities)

	c.HTML(http.StatusOK, "employee_edit", gin.H{
		"title":       "Edit Employee",
		"activePage":  "employee",
		"user":        user,
		"data":        employee,
		"staffing":    staffing,
		"identity":    identity,
		"contact":     contact,
		"address":     address,
		"bankAccount": bankAccount,
		"bloods":      bloods,
		"religions":   religions,
		"banks":       banks,
		"identities":  identities,
		"action":      "/employee/" + id,
		"method":      "POST",
		"isEdit":      true,
		"dateOfBirth": employee.DateOfBirth.Format("2006-01-02"), // jika menggunakan tipe time.Time
		"joinDate":    employee.JoinDate.Format("2006-01-02"),    // jika menggunakan tipe time.Time
	})

}

// POST /departments/:id
func (dc *EmployeeController) Update(c *gin.Context) {
	// Ambil user dari session
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	var employee models.Employee
	if err := config.DB.Where("id = ?", id).First(&employee).Error; err != nil {
		c.String(http.StatusNotFound, "Employee not found")
		return
	}

	// --- Ambil relasi ---
	var staffing models.Staffing
	var bankAccount models.BankAccount
	var identity models.Identity
	var contact models.Contact
	var address models.Address

	if employee.IDStaffing != nil {
		_ = config.DB.First(&staffing, "id = ?", employee.IDStaffing).Error
	}
	if employee.IDBankAccount != nil {
		_ = config.DB.First(&bankAccount, "id = ?", employee.IDBankAccount).Error
	}
	if employee.IDIdentity != nil {
		_ = config.DB.First(&identity, "id = ?", employee.IDIdentity).Error
	}
	if employee.IDContact != nil {
		_ = config.DB.First(&contact, "id = ?", employee.IDContact).Error
	}
	if contact.IDAddress != nil {
		_ = config.DB.First(&address, "id = ?", contact.IDAddress).Error
	}

	// --- Ambil dropdown (selalu dipakai GET/POST) ---
	var bloods []models.Blood
	var religions []models.Religion
	var banks []models.TypeBank
	var identities []models.TypeIdentity

	config.DB.Order("created_at desc").Find(&bloods)
	config.DB.Order("created_at desc").Find(&religions)
	config.DB.Order("created_at desc").Find(&banks)
	config.DB.Order("created_at desc").Find(&identities)

	// Format tanggal untuk input date
	dateOfBirthStr := ""
	joinDateStr := ""
	if !employee.DateOfBirth.IsZero() {
		dateOfBirthStr = employee.DateOfBirth.Format("2006-01-02")
	}
	if !employee.JoinDate.IsZero() {
		joinDateStr = employee.JoinDate.Format("2006-01-02")
	}

	// --- Jika GET: render form edit dan stop ---
	if c.Request.Method == http.MethodGet {
		c.HTML(http.StatusOK, "employee_edit", gin.H{
			"title":       "Edit Employee",
			"action":      "/employee/" + id,
			"data":        employee,
			"staffing":    staffing,
			"bankAccount": bankAccount,
			"identity":    identity,
			"address":     address,
			"contact":     contact,
			"bloods":      bloods,
			"religions":   religions,
			"banks":       banks,
			"identities":  identities,
			"errors":      map[string]string{},
			"form":        map[string]string{},
			"dateOfBirth": dateOfBirthStr,
			"joinDate":    joinDateStr,
		})
		return
	}

	// --- Kalau POST: proses submit ---
	form := map[string]string{
		"id_employee":            c.PostForm("id_employee"),
		"name":                   c.PostForm("name"),
		"email":                  c.PostForm("email"),
		"gender":                 c.PostForm("gender"),
		"citizenship":            c.PostForm("citizenship"),
		"marital_status":         c.PostForm("marital_status"),
		"blood_type":             c.PostForm("blood_type"),
		"religion":               c.PostForm("religion"),
		"work_email":             c.PostForm("work_email"),
		"place_of_birth":         c.PostForm("place_of_birth"),
		"date_of_birth":          c.PostForm("date_of_birth"),
		"join_date":              c.PostForm("join_date"),
		"id_type_bank":           c.PostForm("id_type_bank"),
		"no_account":             c.PostForm("no_account"),
		"no_kpj_bpjs":            c.PostForm("no_kpj_bpjs"),
		"no_bpjs_kes":            c.PostForm("no_bpjs_kes"),
		"no_bpjs_jk":             c.PostForm("no_bpjs_jk"),
		"no_npwp_limabelas":      c.PostForm("no_npwp_limabelas"),
		"no_npwp_enambelas":      c.PostForm("no_npwp_enambelas"),
		"ptkp":                   c.PostForm("ptkp"),
		"id_type_identity":       c.PostForm("id_type_identity"),
		"no":                     c.PostForm("no"),
		"address_identity":       c.PostForm("address_identity"),
		"address_domicile":       c.PostForm("address_domicile"),
		"no_hp":                  c.PostForm("no_hp"),
		"no_emergency_contact":   c.PostForm("no_emergency_contact"),
		"emergency_contact_name": c.PostForm("emergency_contact_name"),
		"emergency_relation":     c.PostForm("emergency_relation"),
	}

	errors := map[string]string{}

	// Validasi photo (sama seperti sebelumnya)
	file, fileErr := c.FormFile("photo")
	if fileErr == nil {
		ct := file.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "image/") {
			errors["photo"] = "Photo must be image (image/*)"
		}
		if file.Size > 2*1024*1024 {
			errors["photo"] = "Max Size 2MB"
		}
	}

	// --- Parse UUID ---
	idBlood, err := parseUUIDPtr(form["blood_type"])
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

	// --- Parse tanggal ---
	var dateOfBirth time.Time
	if form["date_of_birth"] != "" {
		dateOfBirth, err = time.Parse("2006-01-02", form["date_of_birth"])
		if err != nil {
			errors["date_of_birth"] = "Date of birth format not valid"
		}
	}
	var joinDate time.Time
	if form["join_date"] != "" {
		joinDate, err = time.Parse("2006-01-02", form["join_date"])
		if err != nil {
			errors["join_date"] = "Join date format not valid"
		}
	}

	// --- Marital status (string -> bool) ---
	var maritalStatus bool
	switch form["marital_status"] {
	case "":
		errors["marital_status"] = "Marital status is required"
	case "Married":
		maritalStatus = true
	case "Not Married":
		maritalStatus = false
	default:
		errors["marital_status"] = "Invalid marital status value"
	}

	// --- Validasi required basic ---
	if form["name"] == "" {
		errors["name"] = "Full name is required"
	}
	if form["email"] == "" {
		errors["email"] = "Email is required"
	}
	if form["work_email"] == "" {
		errors["work_email"] = "Work email is required"
	}
	if form["gender"] == "" {
		errors["gender"] = "Gender is required"
	}
	if form["citizenship"] == "" {
		errors["citizenship"] = "Citizenship is required"
	}
	if form["place_of_birth"] == "" {
		errors["place_of_birth"] = "Place of birth is required"
	}
	if form["id_employee"] == "" {
		errors["id_employee"] = "ID Employee is required"
	}

	// --- Jika error: render kembali template (TANPA fungsi helper) ---
	if len(errors) > 0 {
		// pakai value yang sudah di‑POST + data lama supaya select tetap terisi
		c.HTML(http.StatusBadRequest, "employee_edit", gin.H{
			"title":       "Edit Employee",
			"action":      "/employee/" + id,
			"data":        employee,
			"staffing":    staffing,
			"bankAccount": bankAccount,
			"identity":    identity,
			"address":     address,
			"contact":     contact,
			"bloods":      bloods,
			"religions":   religions,
			"banks":       banks,
			"identities":  identities,
			"errors":      errors,
			"form":        form,
			"dateOfBirth": form["date_of_birth"], // pakai input user
			"joinDate":    form["join_date"],
		})
		return
	}

	// Kalau tidak ada error, lanjutkan update
	// Upload photo jika ada
	photoURL := employee.Photo // Jangan update photo jika tidak ada file baru
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

	// --- Update Employee ---
	employee.IDEmployee = form["id_employee"]
	employee.Name = form["name"]
	employee.Email = form["email"]
	employee.WorkEmail = form["work_email"]
	employee.Gender = form["gender"]
	employee.Citizenship = form["citizenship"]
	employee.MaritalStatus = maritalStatus
	employee.IDBlood = idBlood
	employee.IDReligion = idReligion
	employee.PlaceOfBirth = form["place_of_birth"]
	employee.DateOfBirth = dateOfBirth
	employee.JoinDate = joinDate
	employee.Photo = photoURL

	// --- Update / create Staffing ---
	if employee.IDStaffing == nil {
		staffing = models.Staffing{}
	}
	staffing.NoBpjsKes = form["no_bpjs_kes"]
	staffing.NoBpjsJk = form["no_bpjs_jk"]
	staffing.NoKpjBpjs = form["no_kpj_bpjs"]
	staffing.NoNpwpLimabelas = form["no_npwp_limabelas"]
	staffing.NoNpwpEnambelas = form["no_npwp_enambelas"]
	staffing.Ptkp = form["ptkp"]
	if err := config.DB.Save(&staffing).Error; err != nil {
		c.String(http.StatusInternalServerError, "update staffing failed: "+err.Error())
		return
	}
	employee.IDStaffing = &staffing.ID

	// --- Update / create BankAccount ---
	if employee.IDBankAccount == nil {
		bankAccount = models.BankAccount{}
	}
	bankAccount.IDTypeBank = idTypeBank
	bankAccount.NoAccount = form["no_account"]
	if err := config.DB.Save(&bankAccount).Error; err != nil {
		c.String(http.StatusInternalServerError, "update bank account failed: "+err.Error())
		return
	}
	employee.IDBankAccount = &bankAccount.ID

	// --- Update / create Identity ---
	if employee.IDIdentity == nil {
		identity = models.Identity{}
	}
	identity.IDTypeIdentity = idTypeIdentity
	identity.No = form["no"]
	if identity.EvidenceLink == "" {
		identity.EvidenceLink = ""
	}
	if err := config.DB.Save(&identity).Error; err != nil {
		c.String(http.StatusInternalServerError, "update identity failed: "+err.Error())
		return
	}
	employee.IDIdentity = &identity.ID

	// --- Update / create Address ---
	if contact.IDAddress == nil {
		address = models.Address{}
	}
	address.AddressIdentity = form["address_identity"]
	address.AddressDomicile = form["address_domicile"]
	if err := config.DB.Save(&address).Error; err != nil {
		c.String(http.StatusInternalServerError, "update address failed: "+err.Error())
		return
	}
	contact.IDAddress = &address.ID

	// --- Update / create Contact ---
	if employee.IDContact == nil {
		contact.ID = uuid.UUID{} // biar gorm generate
	}
	contact.NoHp = form["no_hp"]
	contact.EmergencyContactName = form["emergency_contact_name"]
	contact.EmergencyRelation = form["emergency_relation"]
	contact.NoEmergencyContact = form["no_emergency_contact"]
	if err := config.DB.Save(&contact).Error; err != nil {
		c.String(http.StatusInternalServerError, "update contact failed: "+err.Error())
		return
	}
	employee.IDContact = &contact.ID

	// --- Simpan employee ---
	if err := config.DB.Save(&employee).Error; err != nil {
		c.String(http.StatusInternalServerError, "update employee failed: "+err.Error())
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Employee success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/employee")
}

func (dc *EmployeeController) Delete(c *gin.Context) {
	id := c.Param("id")

	// Cek login (opsional, sama seperti Update)
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.String(http.StatusUnauthorized, "unauthorized")
		return
	}

	var employee models.Employee
	if err := config.DB.Where("id = ?", id).First(&employee).Error; err != nil {
		c.String(http.StatusNotFound, "Employee not found")
		return
	}

	// Mulai transaksi supaya konsisten
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.String(http.StatusInternalServerError, "failed to start transaction")
		return
	}

	// 1. Hapus data yang punya IDEmployee = employee.ID
	if err := tx.Delete(&models.Family{}, "id_employee = ?", employee.ID).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "delete family failed: "+err.Error())
		return
	}

	if err := tx.Delete(&models.Education{}, "id_employee = ?", employee.ID).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "delete education failed: "+err.Error())
		return
	}

	if err := tx.Delete(&models.Career{}, "id_employee = ?", employee.ID).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "delete career failed: "+err.Error())
		return
	}

	if err := tx.Delete(&models.Achievement{}, "id_employee = ?", employee.ID).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "delete achievement failed: "+err.Error())
		return
	}

	// Hapus relasi satu per satu (jika ada)
	// 1. Contact + Address
	if employee.IDContact != nil {
		var contact models.Contact
		if err := tx.First(&contact, "id = ?", employee.IDContact).Error; err == nil {
			// kalau ada address
			if contact.IDAddress != nil {
				if err := tx.Delete(&models.Address{}, "id = ?", contact.IDAddress).Error; err != nil {
					tx.Rollback()
					c.String(http.StatusInternalServerError, "delete address failed: "+err.Error())
					return
				}
			}
			if err := tx.Delete(&contact).Error; err != nil {
				tx.Rollback()
				c.String(http.StatusInternalServerError, "delete contact failed: "+err.Error())
				return
			}
		}
	}

	// 2. Staffing
	if employee.IDStaffing != nil {
		if err := tx.Delete(&models.Staffing{}, "id = ?", employee.IDStaffing).Error; err != nil {
			tx.Rollback()
			c.String(http.StatusInternalServerError, "delete staffing failed: "+err.Error())
			return
		}
	}

	// 3. BankAccount
	if employee.IDBankAccount != nil {
		if err := tx.Delete(&models.BankAccount{}, "id = ?", employee.IDBankAccount).Error; err != nil {
			tx.Rollback()
			c.String(http.StatusInternalServerError, "delete bank account failed: "+err.Error())
			return
		}
	}

	// 4. Identity
	if employee.IDIdentity != nil {
		if err := tx.Delete(&models.Identity{}, "id = ?", employee.IDIdentity).Error; err != nil {
			tx.Rollback()
			c.String(http.StatusInternalServerError, "delete identity failed: "+err.Error())
			return
		}
	}

	// 5. Terakhir, hapus Employee
	if err := tx.Delete(&employee).Error; err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "delete employee failed: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.String(http.StatusInternalServerError, "commit failed: "+err.Error())
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Employee success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/employee")
}
