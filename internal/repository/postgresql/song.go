package postgresql

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SongRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewSongRepository(db *gorm.DB, logger *logrus.Logger) *SongRepository {
	return &SongRepository{
		db:     db,
		logger: logger,
	}
}
