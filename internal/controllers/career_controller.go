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
	})
}

// POST /achievement/:id
func (dc *CareerController) Update(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		ID                string `form:"id" binding:"required"`
		IDEmployee        string `form:"id_employee" binding:"required"`
		IDTypeAchievement string `form:"id_type_achievement" binding:"required"`
		Date              string `form:"date" binding:"required"`
		Title             string `form:"title" binding:"required"`
		Description       string `form:"description"`
		EvidenceLink      string `form:"evidence_link" binding:"required"`
	}

	// Bind form data
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Parse achievement ID
	achID, err := uuid.Parse(input.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid achievement ID",
		})
		return
	}

	// Parse employee ID
	empID, err := uuid.Parse(input.IDEmployee)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid employee ID",
		})
		return
	}

	// Parse type achievement ID
	typeID, err := uuid.Parse(input.IDTypeAchievement)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid type achievement ID",
		})
		return
	}

	// Parse date (format: YYYY-MM-DD dari input type="date")
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid date format",
		})
		return
	}

	// Validate: date tidak boleh di masa depan
	if date.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Date cannot be in the future",
		})
		return
	}

	// Cari achievement yang akan diupdate
	var achievement models.Achievement
	if err := config.DB.First(&achievement, "id = ?", achID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Achievement not found",
		})
		return
	}

	// Update semua fields
	achievement.IDEmployee = &empID
	achievement.IDTypeAchievement = &typeID
	achievement.Date = date
	achievement.Title = input.Title
	achievement.Description = input.Description
	achievement.EvidenceLink = input.EvidenceLink

	// Save ke database
	if err := config.DB.Save(&achievement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update achievement: " + err.Error(),
		})
		return
	}

	// Success response (untuk AJAX)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Achievement success edited",
		"data":    achievement,
	})

	c.Redirect(http.StatusFound, "/achievement/"+id+"/show")
}

// DELETE /achievement/delete/:id - Delete achievement via AJAX
func (ac *CareerController) Delete(c *gin.Context) {
	id := c.Param("id")

	// Parse UUID
	achID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid achievement ID",
		})
		return
	}

	// Cek apakah achievement exists
	var achievement models.Achievement
	if err := config.DB.First(&achievement, "id = ?", achID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Achievement not found",
		})
		return
	}

	// Simpan employee ID sebelum delete
	employeeID := achievement.IDEmployee

	// Cek berapa total achievement untuk employee ini SEBELUM delete
	var totalCount int64
	config.DB.Model(&models.Achievement{}).
		Where("id_employee = ?", employeeID).
		Count(&totalCount)

	// Delete dari database
	if err := config.DB.Delete(&achievement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete achievement: " + err.Error(),
		})
		return
	}

	// Jika ini row terakhir (count = 1 sebelum delete, jadi 0 setelah delete)
	if totalCount == 1 {
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"message":     "Achievement success deleted",
			"redirect":    "/achievement", // ✅ Flag untuk redirect
			"is_last_row": true,
		})
		return
	}

	// Jika masih ada achievement lain untuk employee ini
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Achievement success deleted",
		"redirect":    "", // Kosong = reload halaman sama
		"is_last_row": false,
	})
}

func (ac *CareerController) Excel(ctx *gin.Context) {
	employeeID := ctx.Param("employee_id")

	// Get achievements
	var achievements []models.Achievement
	if err := config.DB.Where("id_employee = ?", employeeID).
		Order("date ASC").
		Find(&achievements).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(achievements) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No achievement found"})
		return
	}

	// Get employee
	var employee models.Employee
	if err := config.DB.Where("id = ?", employeeID).First(&employee).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Get type names untuk semua achievements
	typeNames := make(map[string]string)
	for _, achievement := range achievements {
		idType := achievement.IDTypeAchievement.String()
		if _, exists := typeNames[idType]; !exists {
			var typeAch models.TypeAchievement
			if err := config.DB.Where("id = ?", achievement.IDTypeAchievement).First(&typeAch).Error; err == nil {
				typeNames[idType] = typeAch.Type
			}
		}
	}

	// Create Excel
	f := excelize.NewFile()
	defer f.Close()

	// Hapus sheet default
	f.DeleteSheet("Sheet1")

	// Create sheet
	sheetName := "Achievement"
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
	f.MergeCell(sheetName, "A1", "E1")
	f.SetCellValue(sheetName, "A1", "ACHIEVEMENT REPORT")
	f.SetCellStyle(sheetName, "A1", "E1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)

	// Employee Info
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Employee: %s", employee.Name))
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006 15:04")))

	// Table Headers
	headers := []string{"No", "Date", "Type", "Title", "Description"}
	headerCells := []string{"A", "B", "C", "D", "E"}
	for col, header := range headers {
		cell := fmt.Sprintf("%s%d", headerCells[col], 5)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 5, 20)

	// Data rows
	for idx, achievement := range achievements {
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

		// Date
		cellB := fmt.Sprintf("B%d", row)
		f.SetCellValue(sheetName, cellB, achievement.Date.Format("02 Jan 2006"))
		f.SetCellStyle(sheetName, cellB, cellB, centerStyle)

		// Type
		cellC := fmt.Sprintf("C%d", row)
		typeName := typeNames[achievement.IDTypeAchievement.String()]
		f.SetCellValue(sheetName, cellC, typeName)
		f.SetCellStyle(sheetName, cellC, cellC, cellStyle)

		// Title
		cellD := fmt.Sprintf("D%d", row)
		f.SetCellValue(sheetName, cellD, achievement.Title)
		f.SetCellStyle(sheetName, cellD, cellD, cellStyle)

		// Description
		cellE := fmt.Sprintf("E%d", row)
		f.SetCellValue(sheetName, cellE, achievement.Description)
		f.SetCellStyle(sheetName, cellE, cellE, cellStyle)

		f.SetRowHeight(sheetName, row, 30)
	}

	// Column widths
	f.SetColWidth(sheetName, "A", "A", 5)
	f.SetColWidth(sheetName, "B", "B", 12)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 40)

	// Summary
	summaryRow := len(achievements) + 7
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "Total Achievement:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), len(achievements))

	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("B%d", summaryRow), summaryStyle)

	// Download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=achievement_%s.xlsx", employee.Name))
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	if err := f.Write(ctx.Writer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func (ac *CareerController) Pdf(ctx *gin.Context) {
	employeeID := ctx.Param("employee_id")

	var achievements []models.Achievement
	if err := config.DB.Where("id_employee = ?", employeeID).
		Order("date ASC").
		Find(&achievements).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(achievements) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No achievement found"})
		return
	}

	var employee models.Employee
	if err := config.DB.Where("id = ?", employeeID).First(&employee).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Get type names untuk semua achievements
	typeNames := make(map[string]string) // map[id_type]name
	for _, achievement := range achievements {
		idType := achievement.IDTypeAchievement.String()
		if _, exists := typeNames[idType]; !exists {
			var typeAch models.TypeAchievement
			// Fix: Gunakan string dengan quote untuk UUID
			if err := config.DB.Where("id = ?", idType).First(&typeAch).Error; err == nil {
				typeNames[idType] = typeAch.Type
			}
		}
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "ACHIEVEMENT REPORT", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Employee
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(0, 6, fmt.Sprintf("Employee: %s", employee.Name), "", 1, "L", false, 0, "")

	// Date
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated: %s", time.Now().Format("02 Jan 2006 15:04")), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(5)

	// Group achievements by year
	achievementsByYear := make(map[int][]models.Achievement)
	for _, achievement := range achievements {
		year := achievement.Date.Year()
		achievementsByYear[year] = append(achievementsByYear[year], achievement)
	}

	// Sort years in ascending order
	var years []int
	for year := range achievementsByYear {
		years = append(years, year)
	}
	sort.Ints(years)

	// Render table for each year (ditumpuk dalam 1 page)
	for yearIndex, year := range years {
		if yearIndex > 0 {
			pdf.Ln(5) // Spacing antar tahun
		}

		// Year Header
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("Year %d", year), "", 1, "L", false, 0, "")
		pdf.Ln(2)

		// Table Header
		pdf.SetFont("Arial", "B", 9)
		pdf.SetFillColor(79, 129, 189)
		pdf.SetTextColor(255, 255, 255)

		pdf.CellFormat(8, 7, "No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "Type", "1", 0, "C", true, 0, "")
		pdf.CellFormat(40, 7, "Title", "1", 0, "C", true, 0, "")
		pdf.CellFormat(60, 7, "Description", "1", 0, "C", true, 0, "")
		pdf.CellFormat(52, 7, "Date", "1", 1, "C", true, 0, "")

		// Table Content
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFillColor(242, 242, 242)

		yearAchievements := achievementsByYear[year]
		for idx, achievement := range yearAchievements {
			fill := idx%2 == 1

			// Format date dengan hari
			dateStr := achievement.Date.Format("Monday, 02 January 2006")

			// Get type name from map
			typeName := typeNames[achievement.IDTypeAchievement.String()]

			pdf.CellFormat(8, 7, fmt.Sprintf("%d", idx+1), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(52, 7, dateStr, "1", 1, "C", fill, 0, "")
			pdf.CellFormat(30, 7, typeName, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(40, 7, achievement.Title, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(60, 7, achievement.Description, "1", 0, "L", fill, 0, "")
		}

		pdf.Ln(2)

		// Year Summary
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, fmt.Sprintf("Total Achievement (%d): %d", year, len(yearAchievements)), "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)

	// Total Summary
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Achievement (All Years): %d", len(achievements)), "", 1, "L", false, 0, "")

	// Output ke response writer
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=achievement_%s.pdf", employee.Name))
	ctx.Header("Content-Type", "application/pdf")

	if err := pdf.Error(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err := pdf.Output(ctx.Writer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
