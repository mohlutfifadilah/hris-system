package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-gonic/gin"
)

type ProfileController struct{}

func NewProfileController() *ProfileController {
	return &ProfileController{}
}

func (pc *ProfileController) Index(c *gin.Context) {
	// 1. Ambil user yang sedang login
	currentUser := auth.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 2. Ambil row employee lengkap
	var employee models.Employee
	if err := config.DB.Where("id = ?", currentUser.ID).First(&employee).Error; err != nil {
		c.String(http.StatusInternalServerError, "employee not found")
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

	// 6. Kirim semua ke template profile
	c.HTML(http.StatusOK, "profile", gin.H{
		"title":      "Profile",
		"activePage": "profile",

		"user":         employee,
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

		"educations":   educations,
		"careers":      careers,
		"achievements": achievements,
		"families":     families,

		// master untuk mapping ID -> nama (dipakai di range di template)
		"statuses":        statuses,
		"gradings":        gradings,
		"departments":     departments,
		"familyReligions": familyReligions,
	})
}

// UploadProfilePhoto untuk upload foto profile
func (dc *ProfileController) UploadProfilePhoto(c *gin.Context) {
	userID := auth.GetCurrentUser(c)

	// Ambil file dari form
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan"})
		return
	}

	// Validasi ekstensi file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file not relevant. Use JPG or PNG"})
		return
	}

	// Validasi ukuran file (max 2MB)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size max 2MB"})
		return
	}

	// Get employee data
	var user models.Employee
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hapus foto lama jika ada
	if user.Photo != "" && user.Photo != "static/assets/img/avatar/avatar-1.png" {
		oldPhotoPath := user.Photo
		if _, err := os.Stat(oldPhotoPath); err == nil {
			os.Remove(oldPhotoPath)
		}
	}

	// Generate nama file unik
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("profile_%d_%d%s", userID, timestamp, ext)

	// Path untuk menyimpan file
	uploadDir := "static/assets/img/profiles"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Directory"})
		return
	}

	filepath := filepath.Join(uploadDir, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Upload File"})
		return
	}

	// Update database
	user.Photo = filepath
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Update Database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto Profile success upload",
		"photo":   filepath,
	})
}

// UpdateProfilePhoto untuk edit foto profile
func (dc *ProfileController) UpdateProfilePhoto(c *gin.Context) {
	// Sama seperti UploadProfilePhoto, karena fungsinya mengganti foto
	dc.UploadProfilePhoto(c)
}

// DeleteProfilePhoto untuk hapus foto profile
func (dc *ProfileController) DeleteProfilePhoto(c *gin.Context) {
	userID := auth.GetCurrentUser(c)

	var user models.Employee
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Hapus file foto jika ada
	if user.Photo != "" && user.Photo != "static/assets/img/avatar/avatar-1.png" {
		if _, err := os.Stat(user.Photo); err == nil {
			if err := os.Remove(user.Photo); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting file"})
				return
			}
		}
	}

	// Set ke default avatar
	user.Photo = "static/assets/img/avatar/avatar-1.png"
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error update database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto profile success deleted",
		"photo":   user.Photo,
	})
}
