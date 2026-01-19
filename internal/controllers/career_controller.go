package controllers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CareerController struct{}

func NewCareerController() *CareerController {
	return &CareerController{}
}

// Index - Tampilkan halaman profile
func (dc *CareerController) Index(c *gin.Context) {
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

	var career []models.Career

	err := config.DB.
		Table("career_history").
		Select("DISTINCT ON (id_employee) *").
		Order("id_employee, created_at DESC").
		Scan(&career).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	type EmployeeInfo struct {
		Name      string
		Gender    string
		WorkEmail string
	}
	employeeMap := make(map[string]EmployeeInfo)

	for _, ach := range career { // ganti `emp` jadi `ach`
		if ach.IDEmployee != nil {
			var employee models.Employee
			if err := config.DB.First(&employee, "id = ?", ach.IDEmployee).Error; err == nil {
				employeeMap[ach.IDEmployee.String()] = EmployeeInfo{
					Name:      employee.Name,
					Gender:    employee.Gender, // atau field gender di model Employee kamu
					WorkEmail: employee.WorkEmail,
				}
			}
		}
	}

	statusMap := make(map[string]string)

	for _, car := range career { // ganti `emp` jadi `ach`
		if car.IDStatus != nil {
			var status models.Status
			if err := config.DB.First(&status, "id = ?", car.IDStatus).Error; err == nil {
				statusMap[car.IDStatus.String()] = status.Status // key: string, value: nama employee
			}
		}
	}

	gradingMap := make(map[string]string)

	for _, car := range career { // ganti `emp` jadi `ach`
		if car.IDGrading != nil {
			var grading models.Grading
			if err := config.DB.First(&grading, "id = ?", car.IDGrading).Error; err == nil {
				gradingMap[car.IDGrading.String()] = grading.Grading // key: string, value: nama employee
			}
		}
	}

	departmentMap := make(map[string]string)

	for _, car := range career { // ganti `emp` jadi `ach`
		if car.IDDepartment != nil {
			var department models.Department
			if err := config.DB.First(&department, "id = ?", car.IDDepartment).Error; err == nil {
				departmentMap[car.IDDepartment.String()] = department.Department // key: string, value: nama employee
			}
		}
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render career menggunakan layout main.html
	c.HTML(http.StatusOK, "career", gin.H{
		"title":         "Career",
		"user":          employee,      // seluruh row employee yang login (boleh nil)
		"career":        career,        // seluruh row employee yang login (boleh nil)
		"employeeMap":   employeeMap,   // seluruh row employee yang login (boleh nil)
		"gradingMap":    gradingMap,    // seluruh row employee yang login (boleh nil)
		"departmentMap": departmentMap, // seluruh row employee yang login (boleh nil)
		"statusMap":     statusMap,     // seluruh row employee yang login (boleh nil)
		"activePage":    "career",
		"success":       success,
	})
}

// GET /career/create
func (dc *CareerController) Create(c *gin.Context) {

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

	var employees []models.Employee
	if err := config.DB.Order("created_at desc").Find(&employees).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

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

	c.HTML(http.StatusOK, "career_add", gin.H{
		"title":      "Add Career",
		"activePage": "career",
		"form":       map[string]string{},
		"errors":     map[string]string{},
		"employees":  employees,
		"status":     status,
		"grading":    grading,
		"department": department,
		"action":     "/career",
		"method":     "POST",
		"user":       employee, // seluruh row employee yang login (boleh nil)
	})
}

// POST /career
func (dc *CareerController) Store(c *gin.Context) {

	form := map[string]string{
		"employee":      c.PostForm("employee"),
		"status":        c.PostForm("status"),
		"grading":       c.PostForm("grading"),
		"department":    c.PostForm("department"),
		"effectivedate": c.PostForm("effectivedate"),
		"position":      c.PostForm("position"),
		"evidence_link": c.PostForm("evidence_link"),
	}

	errors := map[string]string{}

	// Parse UUID (return pointer)
	employee, err := parseUUIDPtr(form["employee"])
	if err != nil {
		errors["employee"] = "Employee UUID not valid"
	}

	status, err := parseUUIDPtr(form["status"])
	if err != nil {
		errors["status"] = "Status UUID not valid"
	}

	grading, err := parseUUIDPtr(form["grading"])
	if err != nil {
		errors["grading"] = "Grading UUID not valid"
	}

	department, err := parseUUIDPtr(form["department"])
	if err != nil {
		errors["department"] = "Department UUID not valid"
	}

	// Parse date
	var date time.Time
	if form["effectivedate"] != "" {
		date, err = time.Parse("2006-01-02", form["effectivedate"])
		if err != nil {
			errors["effectivedate"] = "Effective Date format not valid"
		}
	}

	// Kalau ada error, render ulang
	if len(errors) > 0 {
		var employees []models.Employee
		var status []models.Status
		var grading []models.Grading
		var department []models.Department

		config.DB.Order("created_at desc").Find(&employees)
		config.DB.Order("created_at desc").Find(&status)
		config.DB.Order("created_at desc").Find(&grading)
		config.DB.Order("created_at desc").Find(&department)

		c.HTML(http.StatusBadRequest, "career_add", gin.H{
			"title":      "Add Career",
			"action":     "/career",
			"form":       form,
			"errors":     errors,
			"swalError":  "There are some invalid inputs.",
			"employees":  employees,
			"status":     status,
			"grading":    grading,
			"department": department,
			"method":     "POST"})
		return
	}

	// ========== START TRANSACTION ==========
	tx := config.DB.Begin()

	// Create career
	emp := models.Career{
		IDEmployee:    employee,
		IDStatus:      status,
		IDGrading:     grading,
		IDDepartment:  department,
		EffectiveDate: date,
		Position:      form["position"],
		EvidenceLink:  form["evidence_link"],
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := tx.Create(&emp).Error; err != nil {
		c.String(http.StatusInternalServerError, "create career failed: "+err.Error())
		return
	}

	// ========== COMMIT TRANSACTION ==========
	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		c.HTML(http.StatusInternalServerError, "career_add", gin.H{
			"title":  "Add Career",
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

	c.Redirect(http.StatusFound, "/career")
}

// Struct untuk grouping (tambahkan di bagian atas file atau di models)
type CareerYearGroup struct {
	Year    int
	Careers []models.Career
}

// Get show/{id}/career
func (dc *CareerController) Show(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	var employe models.Employee
	if err := config.DB.
		Where("id = ?", currentUser.ID).
		First(&employe).Error; err != nil {
		// handle error (404, dll)
		c.String(http.StatusInternalServerError, "employee not found")
		return
	}

	id := c.Param("id")

	var employee models.Employee
	if err := config.DB.First(&employee, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Employee not found")
		return
	}

	// 1. Ambil SEMUA career employee (NO YEAR FILTER dulu)
	var allCareer []models.Career
	config.DB.Where("id_employee = ?", id).
		Order("effective_date ASC").
		Find(&allCareer)

	if len(allCareer) == 0 {
		// Fallback: empty timeline
		c.HTML(http.StatusOK, "career_info", gin.H{
			"title":      "Career Info",
			"activePage": "career",
			"employee":   employee,
			"yearGroups": []CareerYearGroup{},
			"user":       employe,
		})
		return
	}

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

	c.HTML(http.StatusOK, "career_info", gin.H{
		"title":           "Career Info",
		"activePage":      "career",
		"employee":        employee,
		"yearGroups":      yearGroups,
		"allCareer":       allCareer, // Backup untuk timeline dots
		"statusNames":     statusNames,
		"gradingNames":    gradingNames,
		"departmentNames": departmentNames,
		"user":            employe, // seluruh row employee yang login (boleh nil)
		"status":          status,
		"grading":         grading,
		"department":      department,
	})
}

// POST /achievement/:id
func (dc *CareerController) Update(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		ID            string `form:"id" binding:"required"`
		IDStatus      string `form:"id_status" binding:"required"`
		IDGrading     string `form:"id_grading" binding:"required"`
		IDDepartment  string `form:"id_department" binding:"required"`
		EffectiveDate string `form:"effectivedate" binding:"required"`
		Position      string `form:"position" binding:"required"`
		EvidenceLink  string `form:"evidence_link" binding:"required"`
	}

	// Bind form data
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Parse career ID
	carID, err := uuid.Parse(input.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid career ID",
		})
		return
	}

	// Parse status ID
	statusID, err := uuid.Parse(input.IDStatus)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid status ID",
		})
		return
	}

	// Parse grading ID
	gradingID, err := uuid.Parse(input.IDGrading)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid grading ID",
		})
		return
	}

	// Parse department ID
	departmentID, err := uuid.Parse(input.IDDepartment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid department ID",
		})
		return
	}

	// Parse Effective Date (format: YYYY-MM-DD dari input type="date")
	date, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid date format",
		})
		return
	}

	// Cari career yang akan diupdate
	var career models.Career
	if err := config.DB.First(&career, "id = ?", carID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Career not found",
		})
		return
	}

	// Update semua fields
	career.IDStatus = &statusID
	career.IDGrading = &gradingID
	career.IDDepartment = &departmentID
	career.EffectiveDate = date
	career.Position = input.Position
	career.EvidenceLink = input.EvidenceLink

	// Save ke database
	if err := config.DB.Save(&career).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update career: " + err.Error(),
		})
		return
	}

	// Success response (untuk AJAX)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Career success edited",
		"data":    career,
	})

	c.Redirect(http.StatusFound, "/career/"+id+"/show")
}

// DELETE /career/delete/:id - Delete career via AJAX
func (ac *CareerController) Delete(c *gin.Context) {
	id := c.Param("id")

	// Parse UUID
	carID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid career ID",
		})
		return
	}

	// Cek apakah career exists
	var career models.Career
	if err := config.DB.First(&career, "id = ?", carID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Career not found",
		})
		return
	}

	// Simpan employee ID sebelum delete
	employeeID := career.IDEmployee

	// Cek berapa total career untuk employee ini SEBELUM delete
	var totalCount int64
	config.DB.Model(&models.Career{}).
		Where("id_employee = ?", employeeID).
		Count(&totalCount)

	// Delete dari database
	if err := config.DB.Delete(&career).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete career: " + err.Error(),
		})
		return
	}

	// Jika ini row terakhir (count = 1 sebelum delete, jadi 0 setelah delete)
	if totalCount == 1 {
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"message":     "Career success deleted",
			"redirect":    "/career", // ✅ Flag untuk redirect
			"is_last_row": true,
		})
		return
	}

	// Jika masih ada career lain untuk employee ini
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Career success deleted",
		"redirect":    "", // Kosong = reload halaman sama
		"is_last_row": false,
	})
}

func (ac *CareerController) Excel(ctx *gin.Context) {
	employeeID := ctx.Param("employee_id")

	// Get career
	var careers []models.Career
	if err := config.DB.Where("id_employee = ?", employeeID).
		Order("effective_date ASC").
		Find(&careers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(careers) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No career found"})
		return
	}

	// Get employee
	var employee models.Employee
	if err := config.DB.Where("id = ?", employeeID).First(&employee).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Get status untuk semua career
	statuss := make(map[string]string)
	for _, career := range careers {
		idType := career.IDStatus.String()
		if _, exists := statuss[idType]; !exists {
			var typeStatus models.Status
			if err := config.DB.Where("id = ?", career.IDStatus).First(&typeStatus).Error; err == nil {
				statuss[idType] = typeStatus.Status
			}
		}
	}

	// Get grading untuk semua career
	gradings := make(map[string]string)
	for _, career := range careers {
		idType := career.IDGrading.String()
		if _, exists := gradings[idType]; !exists {
			var typeGrading models.Grading
			if err := config.DB.Where("id = ?", career.IDGrading).First(&typeGrading).Error; err == nil {
				gradings[idType] = typeGrading.Grading
			}
		}
	}

	// Get department untuk semua career
	departments := make(map[string]string)
	for _, career := range careers {
		idType := career.IDDepartment.String()
		if _, exists := departments[idType]; !exists {
			var typeDepartment models.Department
			if err := config.DB.Where("id = ?", career.IDDepartment).First(&typeDepartment).Error; err == nil {
				departments[idType] = typeDepartment.Department
			}
		}
	}

	// Create Excel
	f := excelize.NewFile()
	defer f.Close()

	// Hapus sheet default
	f.DeleteSheet("Sheet1")

	// Create sheet
	sheetName := "Sheet1"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// ===== STYLES =====

	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// Header style (blue background)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4F81BD"},
			Pattern: 1,
		},
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
			Size:  11,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Data style
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// Data center style
	dataCenterStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// ===== CONTENT =====

	// Title
	f.MergeCell(sheetName, "A1", "F1")
	f.SetCellValue(sheetName, "A1", "CAREER REPORT")
	f.SetCellStyle(sheetName, "A1", "F1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)

	// Employee Info
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Employee: %s", employee.Name))
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006 15:04")))

	// Table Headers
	headers := []string{"No", "Effective Date", "Status", "Grading", "Position", "Department"}
	headerCells := []string{"A", "B", "C", "D", "E", "F"}
	for col, header := range headers {
		cell := fmt.Sprintf("%s%d", headerCells[col], 5)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 5, 20)

	// Data rows
	for idx, career := range careers {
		row := idx + 6
		fill := idx%2 == 1
		cellStyle := dataStyle
		centerStyle := dataCenterStyle
		if fill {
			cellStyle, _ = f.NewStyle(&excelize.Style{
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"F2F2F2"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "left",
					Vertical:   "top",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
				},
			})
			centerStyle, _ = f.NewStyle(&excelize.Style{
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"F2F2F2"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
				},
			})
		}

		// No
		cellA := fmt.Sprintf("A%d", row)
		f.SetCellValue(sheetName, cellA, idx+1)
		f.SetCellStyle(sheetName, cellA, cellA, centerStyle)

		// Effective Date
		cellB := fmt.Sprintf("B%d", row)
		f.SetCellValue(sheetName, cellB, career.EffectiveDate.Format("02 Jan 2006"))
		f.SetCellStyle(sheetName, cellB, cellB, centerStyle)

		// Status
		cellC := fmt.Sprintf("C%d", row)
		status := statuss[career.IDStatus.String()]
		f.SetCellValue(sheetName, cellC, status)
		f.SetCellStyle(sheetName, cellC, cellC, cellStyle)

		// Grading
		cellD := fmt.Sprintf("D%d", row)
		grading := gradings[career.IDGrading.String()]
		f.SetCellValue(sheetName, cellD, grading)
		f.SetCellStyle(sheetName, cellD, cellD, cellStyle)

		// Position
		cellE := fmt.Sprintf("E%d", row)
		f.SetCellValue(sheetName, cellE, career.Position)
		f.SetCellStyle(sheetName, cellE, cellE, cellStyle)

		// Department
		cellF := fmt.Sprintf("F%d", row)
		department := departments[career.IDDepartment.String()]
		f.SetCellValue(sheetName, cellF, department)
		f.SetCellStyle(sheetName, cellF, cellF, cellStyle)

		f.SetRowHeight(sheetName, row, 30)
	}

	// Column widths
	f.SetColWidth(sheetName, "A", "A", 5)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 40)
	f.SetColWidth(sheetName, "F", "F", 40)

	// Download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=career_%s.xlsx", employee.Name))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(ctx.Writer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func (ac *CareerController) Pdf(ctx *gin.Context) {
	employeeID := ctx.Param("employee_id")

	var careers []models.Career
	if err := config.DB.Where("id_employee = ?", employeeID).
		Order("effective_date ASC").
		Find(&careers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(careers) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No career found"})
		return
	}

	var employee models.Employee
	if err := config.DB.Where("id = ?", employeeID).First(&employee).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Get type status
	typeStatuss := make(map[string]string)
	for _, career := range careers {
		if career.IDStatus == nil {
			continue
		}
		idStatus := career.IDStatus.String()
		if _, exists := typeStatuss[idStatus]; !exists {
			var typeCar models.Status
			if err := config.DB.Where("id = ?", idStatus).First(&typeCar).Error; err == nil {
				typeStatuss[idStatus] = typeCar.Status
			}
		}
	}

	// Get type grading
	typeGradings := make(map[string]string)
	for _, career := range careers {
		if career.IDGrading == nil {
			continue
		}
		idGrading := career.IDGrading.String()
		if _, exists := typeGradings[idGrading]; !exists {
			var typeCar models.Grading
			if err := config.DB.Where("id = ?", idGrading).First(&typeCar).Error; err == nil {
				typeGradings[idGrading] = typeCar.Grading
			}
		}
	}

	// Get type department
	typeDepartments := make(map[string]string)
	for _, career := range careers {
		if career.IDDepartment == nil {
			continue
		}
		idDepartment := career.IDDepartment.String()
		if _, exists := typeDepartments[idDepartment]; !exists {
			var typeCar models.Department
			if err := config.DB.Where("id = ?", idDepartment).First(&typeCar).Error; err == nil {
				typeDepartments[idDepartment] = typeCar.Department
			}
		}
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "CAREER REPORT", "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Employee
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Employee: %s", employee.Name), "", 1, "L", false, 0, "")

	// Date
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2026 15:04")), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)

	// Group careers by year
	careersByYear := make(map[int][]models.Career)
	for _, career := range careers {
		year := career.EffectiveDate.Year()
		careersByYear[year] = append(careersByYear[year], career)
	}

	// Sort years
	var years []int
	for year := range careersByYear {
		years = append(years, year)
	}
	sort.Ints(years)

	// Render each year
	for yearIndex, year := range years {
		if yearIndex > 0 {
			pdf.Ln(8)
		}

		// Year Header
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("Year %d", year), "", 1, "L", false, 0, "")
		pdf.Ln(2)

		// Table Header
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(79, 129, 189)
		pdf.SetTextColor(255, 255, 255)

		pdf.CellFormat(10, 7, "No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Date", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "Status", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "Grading", "1", 0, "C", true, 0, "")
		pdf.CellFormat(42, 7, "Position", "1", 0, "C", true, 0, "")
		pdf.CellFormat(38, 7, "Department", "1", 1, "C", true, 0, "")

		// Table Content
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)

		yearCareers := careersByYear[year]
		for idx, career := range yearCareers {
			fill := idx%2 == 0
			if fill {
				pdf.SetFillColor(242, 242, 242)
			}

			dateStr := career.EffectiveDate.Format("02 Jan 2006")
			status := typeStatuss[career.IDStatus.String()]
			grading := typeGradings[career.IDGrading.String()]
			dept := typeDepartments[career.IDDepartment.String()]

			pdf.CellFormat(10, 7, fmt.Sprintf("%d", idx+1), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(40, 7, dateStr, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 7, status, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 7, grading, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(42, 7, career.Position, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(38, 7, dept, "1", 1, "L", fill, 0, "")
		}

		pdf.Ln(2)

		// Year Summary
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, fmt.Sprintf("Total Career (%d): %d", year, len(yearCareers)), "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)

	// Total Summary
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Career (All Years): %d", len(careers)), "", 1, "L", false, 0, "")

	// Output
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=career_%s.pdf", employee.Name))
	ctx.Header("Content-Type", "application/pdf")

	if err := pdf.Output(ctx.Writer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

