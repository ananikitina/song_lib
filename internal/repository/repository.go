package repository

import (
	"github.com/ananikitina/song_lib/internal/models"
)

type SongRepository interface {
	Add(song *models.Song) error
}
