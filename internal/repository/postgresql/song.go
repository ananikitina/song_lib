package postgresql

import (
	"fmt"
	"strings"

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

// Добавление песни
func (r *songRepository) Add(song *models.Song) error {
	if err := r.db.Create(song).Error; err != nil {
		r.logger.Errorf("Add: failed to add song to database: %v", err)
		return err
	}
	r.logger.Infof("Add: song added successfully")
	return nil
}

// Получение всех песен
func (r *songRepository) GetAll() ([]models.Song, error) {
	var songs []models.Song
	res := r.db.Find(&songs)
	if res.Error != nil {
		r.logger.Errorf("GetAll: failed to fetch all songs from database: %v", res.Error)
		return nil, res.Error
	}

	r.logger.Infof("GetAll: successfully fetched %d songs from database", len(songs))
	return songs, nil
}

// Получение песни по ID
func (r *songRepository) GetById(id uint) (*models.Song, error) {
	var song models.Song
	res := r.db.First(&song, id)
	if res.Error != nil {
		r.logger.Errorf("GetById: failed to get song from database with ID %d: %v", id, res.Error)
		return nil, res.Error
	}

	r.logger.Infof("GetById: successfully retrieved song from database with ID %d", id)
	return &song, nil
}

// Получение песни с фильтрацией и пагинацией
func (r *songRepository) GetWithFiltersAndPagination(filters map[string]interface{}, page int, pageSize int) ([]models.Song, error) {
	var songs []models.Song
	query := r.db.Model(&models.Song{})

	r.logger.Debugf("GetWithFiltersAndPagination: SQL query: %v", query.Statement.SQL.String())
	delete(filters, "page")
	delete(filters, "pageSize")
	for key, value := range filters {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	r.logger.Debugf("GetWithFiltersAndPagination: query with filters: %v", query)
	res := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&songs)
	if res.Error != nil {
		r.logger.Errorf("GetWithFiltersAndPagination: failed to fetch songs with filters and pagination: %v", res.Error)
		return nil, res.Error
	}

	r.logger.Infof("GetWithFiltersAndPagination: successfully fetched %d songs with filters and pagination", len(songs))
	return songs, nil
}

// Обновление песни
func (r *songRepository) Update(song *models.Song) error {
	r.logger.Infof("Update: updating song in database with ID %d", song.ID)
	if err := r.db.Save(song).Error; err != nil {
		r.logger.Errorf("Update: failed to update song in database with ID %d: %v", song.ID, err)
		return err
	}

	r.logger.Infof("Update: song with ID %d updated successfully in database", song.ID)
	return nil
}

// Удаление песни
func (r *songRepository) Delete(id uint) error {
	r.logger.Infof("Delete: deleting song from database with ID %d", id)
	if err := r.db.Delete(&models.Song{}, id).Error; err != nil {
		r.logger.Errorf("Delete: failed to delete song from database with ID %d: %v", id, err)
		return err
	}

	r.logger.Infof("Delete: song with ID %d deleted successfully", id)
	return nil
}

// Получение текста песни с пагинацией по куплетам
func (r *songRepository) GetVersesWithPagination(id uint, page int, pageSize int) ([]string, error) {
	var song models.Song
	res := r.db.First(&song, id)
	if res.Error != nil {
		r.logger.Errorf("GetVersesWithPagination: failed to get song from database with ID %d: %v", id, res.Error)
		return nil, res.Error
	}

	verses := strings.Split(song.Text, "\n") // куплеты разделены новой строкой?

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(verses) {
		return []string{}, nil
	}
	if end > len(verses) {
		end = len(verses)
	}

	return verses[start:end], nil
}
