package models

import (
	"time"

	"github.com/google/uuid"
)

type RankHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Rank      string    `form:"rank" binding:"required" gorm:"size:50;not null"` // rank
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RankHistory) TableName() string {
	return "rank_history"
}
