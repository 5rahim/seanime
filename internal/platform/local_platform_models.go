package platform

import "time"

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LocalMedia struct {
	BaseModel
	ID    int    `gorm:"column:id" json:"id"`       // Media ID
	Type  string `gorm:"column:type" json:"type"`   // "anime" or "manga"
	Value []byte `gorm:"column:value" json:"value"` // Marshalled struct
}
