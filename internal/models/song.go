package models

import "time"

type Song struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	GroupName   string    `json:"group" gorm:"column:group_name"`
	SongName    string    `json:"song" gorm:"column:song_name"`
	ReleaseDate string    `json:"releaseDate,omitempty" gorm:"column:release_date"`
	Text        string    `json:"text,omitempty" gorm:"column:text"`
	Link        string    `json:"link,omitempty" gorm:"column:link"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}
