package service

import (
	"github.com/ananikitina/song_lib/internal/models"
)

type SongService interface {
	AddSong(groupName, songName string) (*models.Song, error)
	GetSongInfo(group, song string) (*models.SongDetail, error)
}
