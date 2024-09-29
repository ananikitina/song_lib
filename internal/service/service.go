package service

import (
	"context"

	"github.com/ananikitina/song_lib/internal/models"
)

type SongService interface {
	AddSong(ctx context.Context, groupName, songName string) (*models.Song, error)
	GetSongInfo(ctx context.Context, groupName, songName string) (*models.SongDetail, error)
	GetSongById(id uint) (*models.Song, error)
	GetAllSongs() ([]models.Song, error)
	UpdateSong(songId uint, updatedSong models.Song) (*models.Song, error)
	DeleteSong(id uint) error
	GetSongsWithFiltersAndPagination(filters map[string]interface{}, page int, pageSize int) ([]models.Song, error)
	GetSongVersesWithPagination(songId uint, page int, pageSize int) ([]string, error)
}
