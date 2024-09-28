package repository

import (
	"github.com/ananikitina/song_lib/internal/models"
)

type SongRepository interface {
	Add(song *models.Song) error
	GetAll() ([]models.Song, error)
	GetById(id uint) (*models.Song, error)
	GetWithFiltersAndPagination(filters map[string]interface{}, page int, pageSize int) ([]models.Song, error)
	Update(song *models.Song) error
	Delete(id uint) error
	GetVersesWithPagination(id uint, page int, pageSize int) ([]string, error)
}
