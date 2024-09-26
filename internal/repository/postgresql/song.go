package postgresql

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/ananikitina/song_lib/internal/models"
	"github.com/ananikitina/song_lib/internal/repository"
)

type songRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewSongRepository(db *gorm.DB, logger *logrus.Logger) repository.SongRepository {
	return &songRepository{
		db:     db,
		logger: logger,
	}
}

func (r *songRepository) Add(song *models.Song) error {
	if err := r.db.Create(song).Error; err != nil {
		r.logger.Errorf("failed to add song to database: %v", err)
		return err
	}
	r.logger.Infof("song added successfully")
	return nil
}
