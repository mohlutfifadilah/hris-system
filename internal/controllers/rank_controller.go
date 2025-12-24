package controllers

import (
	"net/http"

	"hris-system/config"
	auth "hris-system/internal/auth"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type RankController struct{}

func NewRankController() *RankController {
	return &RankController{}
}

// Index - Tampilkan halaman
func (dc *RankController) Index(c *gin.Context) {
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

	// // 2) Ambil row career lengkap dari DB
	// var career models.Career
	// if err := config.DB.
	// 	Where("id = ?", employee.IDCareer).
	// 	First(&career).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "career not found")
	// 	return
	// }

	// // 3) Ambil row career_history lengkap dari DB
	// var career_history models.CareerHistory
	// if err := config.DB.
	// 	Where("id = ?", career.IDCareerHistory).
	// 	First(&career_history).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "career history not found")
	// 	return
	// }

	// // 4) Ambil row department_history lengkap dari DB
	// var department_history models.DepartmentHistory
	// if err := config.DB.
	// 	Where("id = ?", career_history.IDDepartment).
	// 	First(&department_history).Error; err != nil {
	// 	// handle error (404, dll)
	// 	c.String(http.StatusInternalServerError, "department history not found")
	// 	return
	// }

	var ranks []models.RankHistory
	if err := config.DB.Order("created_at desc").Find(&ranks).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	session := sessions.Default(c)
	success := session.Get("flash_success")
	if success != nil {
		session.Delete("flash_success")
		_ = session.Save()
	}

	// Render rank menggunakan layout main.html
	c.HTML(http.StatusOK, "rank", gin.H{
		"title":      "Rank",
		"user":       employee, // seluruh row employee yang login (boleh nil)
		"ranks":      ranks,    // seluruh row employee yang login (boleh nil)
		"activePage": "rank",
		"success":    success,
	})
}

// GET /ranks/create
func (dc *RankController) Create(c *gin.Context) {
	c.HTML(http.StatusOK, "rank_add", gin.H{
		"title":      "Add Rank",
		"activePage": "rank",
		"action":     "/ranks",
		"method":     "POST",
	})
}

// POST /ranks
func (dc *RankController) Store(c *gin.Context) {
	var input models.RankHistory

	// ambil field "rank" dari form
	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "rank_add", gin.H{
			"title":      "Add Rank",
			"activePage": "rank",
			"error":      "Name rank required",
			"data":       input,
			"action":     "/ranks",
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
	session.Set("flash_success", "Rank success added")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/rank")
}

// GET /ranks/:id/rank
func (dc *RankController) Edit(c *gin.Context) {
	id := c.Param("id")

	var rank models.RankHistory
	if err := config.DB.First(&rank, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Rank not found")
		return
	}

	c.HTML(http.StatusOK, "rank_edit", gin.H{
		"title":      "Edit Rank",
		"activePage": "rank",
		"data":       rank,
		"action":     "/ranks/" + id,
		"method":     "POST",
		"isEdit":     true,
	})
}

// POST /ranks/:id
func (dc *RankController) Update(c *gin.Context) {
	id := c.Param("id")

	var rank models.RankHistory
	if err := config.DB.First(&rank, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Rank not found")
		return
	}

	var input struct {
		Rank string `form:"rank" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.HTML(http.StatusBadRequest, "rank_edit", gin.H{
			"title":      "Edit Rank",
			"activePage": "rank",
			"error":      "Name rank required",
			"data":       rank,
			"action":     "/ranks/" + id,
			"method":     "POST",
			"isEdit":     true,
		})
		return
	}

	rank.Rank = input.Rank

	if err := config.DB.Save(&rank).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal mengupdate: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Rank success edited")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/rank")
}

func (dc *RankController) Delete(c *gin.Context) {
	id := c.Param("id")

	var rank models.RankHistory
	if err := config.DB.First(&rank, "id = ?", id).Error; err != nil {
		c.String(http.StatusNotFound, "Rank not found")
		return
	}

	if err := config.DB.Delete(&rank).Error; err != nil {
		c.String(http.StatusInternalServerError, "Gagal menghapus: %v", err)
		return
	}

	// set flash
	session := sessions.Default(c)
	session.Set("flash_success", "Rank success deleted")
	_ = session.Save()

	c.Redirect(http.StatusFound, "/rank")
}
