package domain

import (
	"github.com/ananikitina/song_lib/internal/repository"
	"github.com/sirupsen/logrus"
)

type SongService struct {
	songRepo repository.SongRepository
	logger   *logrus.Logger
}

func NewSongService(repo repository.SongRepository, logger *logrus.Logger) *SongService {
	return &SongService{
		songRepo: repo,
		logger:   logger,
	}
}
