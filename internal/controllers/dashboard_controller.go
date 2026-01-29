package controllers

import (
	"fmt"
	"net/http"
	"sort"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-gonic/gin"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// Index - Tampilkan halaman dashboard
func (dc *DashboardController) Index(c *gin.Context) {
	// Ambil user dari session (helper, tanpa middleware)
	currentUser := auth.GetCurrentUser(c) // *models.Employee atau nil

	// Hitung total karyawan
	var totalEmployees int64
	config.DB.Model(&models.Employee{}).Count(&totalEmployees)

	// Hitung total career
	var totalCareers int64
	config.DB.Model(&models.Career{}).Count(&totalCareers)

	// Hitung total achievements
	var totalAchievements int64
	config.DB.Model(&models.Achievement{}).Count(&totalAchievements)

	var totalDepartments int64
	config.DB.Model(&models.Department{}).Count(&totalDepartments)

	var achievements []models.Achievement
	if err := config.DB.
		Order("created_at desc"). // Mengurutkan berdasarkan waktu pembuatan, yang terbaru di atas
		Limit(3).                 // Membatasi hasil hanya 5 data
		Find(&achievements).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error fetching achievements")
		return
	}

	// TypeAchievementMap
	typeAchievementMap := make(map[string]string)

	for _, typ := range achievements {
		if typ.IDTypeAchievement != nil {
			var typeAchievement models.TypeAchievement
			if err := config.DB.First(&typeAchievement, "id = ?", typ.IDTypeAchievement).Error; err == nil {
				typeAchievementMap[typ.IDTypeAchievement.String()] = typeAchievement.Type
			}
		}
	}

	type TopAchievement struct {
		TypeName string `json:"typeName"`
		Count    int    `json:"count"`
		Title    string `json:"title"`
	}

	// 1. Hitung frekuensi tiap TypeAchievement
	typeFreq := make(map[string]int)
	for _, ach := range achievements {
		if ach.IDTypeAchievement != nil && typeAchievementMap[ach.IDTypeAchievement.String()] != "" {
			typeFreq[typeAchievementMap[ach.IDTypeAchievement.String()]]++
		}
	}

	// 2. Buat slice untuk sorting
	var typeAchievements []TopAchievement
	for typeName, count := range typeFreq {
		// Cari 1 achievement dengan type ini untuk contoh title
		var exampleTitle string
		for _, ach := range achievements {
			if typeAchievementMap[ach.IDTypeAchievement.String()] == typeName {
				exampleTitle = ach.Title
				break
			}
		}

		if exampleTitle != "" {
			typeAchievements = append(typeAchievements, TopAchievement{
				TypeName: typeName,
				Count:    count,
				Title:    exampleTitle,
			})
		}
	}

	// 3. Sort DESC by count
	sort.Slice(typeAchievements, func(i, j int) bool {
		return typeAchievements[i].Count > typeAchievements[j].Count
	})

	// 4. Ambil top 5
	topAchievements := typeAchievements
	if len(topAchievements) > 5 {
		topAchievements = topAchievements[:5]
	}

	// EmployeeMap
	type EmployeeInfo struct {
		Name  string
		Photo string
	}
	employeeMap := make(map[string]EmployeeInfo)

	for _, ach := range achievements {
		if ach.IDEmployee != nil {
			var emp models.Employee
			if err := config.DB.First(&emp, "id = ?", ach.IDEmployee).Error; err == nil {
				employeeMap[ach.IDEmployee.String()] = EmployeeInfo{
					Name:  emp.Name,
					Photo: emp.Photo,
				}
			}
		}
	}

	// CareerMap - INI YANG PENTING
	type CareerInfo struct {
		Position   string
		Department string
	}
	careerMap := make(map[string]CareerInfo)

	for _, ach := range achievements {
		if ach.IDEmployee != nil {
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
			`, ach.IDEmployee.String())

			if err := config.DB.Raw(query).Scan(&result).Error; err == nil {
				careerMap[ach.IDEmployee.String()] = CareerInfo{
					Position:   result.Position,
					Department: result.Department,
				}
			}
		}
	}

	// Render dashboard menggunakan layout main.html
	c.HTML(http.StatusOK, "dashboard", gin.H{
		"title":              "Dashboard",
		"user":               currentUser, // seluruh row employee yang login (boleh nil)
		"activePage":         "dashboard",
		"totalEmployees":     totalEmployees,
		"totalDepartments":   totalDepartments,
		"totalCareers":       totalCareers,
		"totalAchievements":  totalAchievements,
		"achievements":       achievements,
		"employeeMap":        employeeMap,
		"typeAchievementMap": typeAchievementMap,
		"careerMap":          careerMap, // PASTIKAN INI ADA
		"topAchievements":    topAchievements,
	})
}
