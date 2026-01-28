package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	// Mengambil data achievement per bulan
	var counts []int64
	currentYear := time.Now().Year()

	// Loop untuk bulan 1 hingga bulan sekarang - 1
	for month := 1; month <= int(time.Now().Month())-1; month++ {
		var count int64
		err := config.DB.
			Model(&models.Achievement{}).
			Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?", month, currentYear).
			Count(&count).Error
		if err != nil {
			c.String(http.StatusInternalServerError, "Error fetching monthly achievement counts")
			return
		}
		counts = append(counts, count)
	}

	// Labels untuk bulan (January, February, ... sampai bulan terakhir)
	labels := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

	// Konversi data counts dan labels ke JSON
	countsJSON, err := json.Marshal(counts)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error converting counts to JSON")
		return
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error converting labels to JSON")
		return
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
		"careerMap":          careerMap,          // PASTIKAN INI ADA
		"monthlyCounts":      string(countsJSON), // Kirim data JSON ke template
		"labels":             string(labelsJSON), // Kirim data JSON ke template
	})
}
