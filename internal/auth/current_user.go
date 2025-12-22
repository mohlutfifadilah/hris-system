package auth

import (
	"hris-system/config"
	"hris-system/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetCurrentUser(c *gin.Context) *models.Employee {
	session := sessions.Default(c)
	v := session.Get("employee_id")
	if v == nil {
		return nil
	}

	idStr, ok := v.(string)
	if !ok {
		return nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil
	}

	var emp models.Employee
	if err := config.DB.First(&emp, "id = ?", id).Error; err != nil {
		return nil
	}
	return &emp
}
