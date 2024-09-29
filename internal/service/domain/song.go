package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ananikitina/song_lib/config"
	"github.com/ananikitina/song_lib/internal/models"
	"github.com/ananikitina/song_lib/internal/repository"
	"github.com/ananikitina/song_lib/internal/service"
	"github.com/sirupsen/logrus"
)

// Кастомные ошибки
var (
	ErrInvalidID        = errors.New("invalid song ID")
	ErrEmptyParameters  = errors.New("parameters must not be empty")
	ErrSongNotFound     = errors.New("song not found")
	ErrFailedAPIRequest = errors.New("failed to fetch data from external API")
)

type songService struct {
	repo   repository.SongRepository
	logger *logrus.Logger
	client *http.Client
}

func NewSongService(repo repository.SongRepository, logger *logrus.Logger) service.SongService {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &songService{
		repo:   repo,
		logger: logger,
		client: client,
	}
}

// Валидация ID
func (s *songService) validateId(id uint) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return nil
}

// Валидация непустых параметров
func (s *songService) validateNonEmptyParams(params ...string) error {
	for _, param := range params {
		if param == "" {
			return ErrEmptyParameters
		}
	}
	return nil
}

// Получение информации о песне из внешнего API
func (s *songService) GetSongInfo(ctx context.Context, groupName, songName string) (*models.SongDetail, error) {
	if err := s.validateNonEmptyParams(groupName, songName); err != nil {
		s.logger.Warn("GetSongInfo: groupName or songName is empty")
		return nil, err
	}

	s.logger.Infof("GetSongInfo: fetching song info for group: %s and song: %s", groupName, songName)

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	encodedGroup := url.QueryEscape(groupName)
	encodedSong := url.QueryEscape(songName)
	url := fmt.Sprintf("%s/info?group=%s&song=%s", cfg.ExternalApi, encodedGroup, encodedSong)

	// Создание запроса с контекстом
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Errorf("GetSongInfo: failed to fetch data from external API: %v", err)
		return nil, ErrFailedAPIRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("GetSongInfo: unexpected status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var songDetail models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		s.logger.Errorf("GetSongInfo: failed to decode external API response: %v", err)
		return nil, fmt.Errorf("failed to decode song info response: %w", err)
	}

	s.logger.Infof("GetSongInfo: successfully fetched song info for group: %s and song: %s", groupName, songName)
	return &songDetail, nil
}

// Добавление песни
func (s *songService) AddSong(ctx context.Context, groupName, songName string) (*models.Song, error) {
	if err := s.validateNonEmptyParams(groupName, songName); err != nil {
		s.logger.Warn("AddSong: groupName or songName is empty")
		return nil, err
	}

	s.logger.Infof("AddSong: creating song: %s with group: %s", songName, groupName)

	// Получение информации о песне через внешний API
	songDetail, err := s.GetSongInfo(ctx, groupName, songName)
	if err != nil {
		s.logger.Errorf("failed to fetch song info: %v", err)
		return nil, err
	}

	// Создание новой записи песни
	song := &models.Song{
		GroupName:   groupName,
		SongName:    songName,
		ReleaseDate: songDetail.ReleaseDate,
		Text:        songDetail.Text,
		Link:        songDetail.Link,
	}

	// Сохранение песни в базе данных
	if err := s.repo.Add(song); err != nil {
		s.logger.Errorf("AddSong: failed to save song to database: %v", err)
		return nil, err
	}

	s.logger.Infof("AddSong: song created with ID: %d", song.ID)
	return song, nil
}

// Получение песни по ID
func (s *songService) GetSongById(id uint) (*models.Song, error) {
	if err := s.validateId(id); err != nil {
		s.logger.Warn("GetSongById: invalid id")
		return nil, err
	}

	s.logger.Infof("GetSongById: getting song with id: %d", id)

	song, err := s.repo.GetById(id)
	if err != nil {
		s.logger.Errorf("GetSongById: failed to fetch song from database: %v", err)
		return nil, ErrSongNotFound
	}

	s.logger.Infof("GetSongById: got song with ID: %d", song.ID)
	return song, nil
}

// Получение всех песен
func (s *songService) GetAllSongs() ([]models.Song, error) {
	s.logger.Info("GetAllSongs: fetching all songs")
	songs, err := s.repo.GetAll()
	if err != nil {
		s.logger.Errorf("GetAllSongs: failed to fetch songs: %v", err)
		return nil, err
	}
	return songs, nil
}

// Вспомогательная функция для обновления полей песни
func updateNonEmptyFields(target, source *models.Song) {
	if source.GroupName != "" {
		target.GroupName = source.GroupName
	}
	if source.SongName != "" {
		target.SongName = source.SongName
	}
	if source.ReleaseDate != "" {
		target.ReleaseDate = source.ReleaseDate
	}
	if source.Text != "" {
		target.Text = source.Text
	}
	if source.Link != "" {
		target.Link = source.Link
	}
}

// Обновление песни
func (s *songService) UpdateSong(songId uint, updatedSong models.Song) (*models.Song, error) {
	if err := s.validateId(songId); err != nil {
		s.logger.Warn("UpdateSong: invalid id")
		return nil, err
	}

	s.logger.Infof("UpdateSong: updating song with ID: %d", songId)

	song, err := s.repo.GetById(songId)
	if err != nil {
		s.logger.Errorf("UpdateSong: failed to update song: %v", err)
		return nil, ErrSongNotFound
	}

	// Обновление полей песни
	updateNonEmptyFields(song, &updatedSong)
	song.UpdatedAt = time.Now()

	if err := s.repo.Update(song); err != nil {
		s.logger.Errorf("UpdateSong: failed to update song: %v", err)
		return nil, err
	}

	s.logger.Infof("UpdateSong: song updated with ID: %d", song.ID)
	return song, nil
}

// Удаление песни
func (s *songService) DeleteSong(id uint) error {
	if err := s.validateId(id); err != nil {
		s.logger.Warn("DeleteSong: invalid id")
		return err
	}

	s.logger.Infof("DeleteSong: deleting song with ID: %d", id)
	if err := s.repo.Delete(id); err != nil {
		s.logger.Errorf("DeleteSong: failed to delete song: %v", err)
		return err
	}
	s.logger.Infof("DeleteSong: song deleted with ID: %d", id)
	return nil
}

// Получение песен с фильтрацией и пагинацией
func (s *songService) GetSongsWithFiltersAndPagination(filters map[string]interface{}, page int, pageSize int) ([]models.Song, error) {
	if page <= 0 || pageSize <= 0 {
		s.logger.Warn("GetSongsWithFiltersAndPagination: page and pageSize must be greater than zero")
		return nil, fmt.Errorf("page and pageSize must be greater than zero")
	}

	s.logger.Infof("GetSongsWithFiltersAndPagination: fetching songs with filters %v, page: %d, pageSize: %d", filters, page, pageSize)

	songs, err := s.repo.GetWithFiltersAndPagination(filters, page, pageSize)
	if err != nil {
		s.logger.Errorf("GetSongsWithFiltersAndPagination: failed to fetch songs with filters: %v", err)
		return nil, err
	}

	return songs, nil
}

// Получение текста песни с пагинацией по куплетам
func (s *songService) GetSongVersesWithPagination(songId uint, page int, pageSize int) ([]string, error) {
	if page <= 0 || pageSize <= 0 {
		s.logger.Warn("GetSongsWithFiltersAndPagination : page and pageSize must be greater than zero")
		return nil, fmt.Errorf("page and pageSize must be greater than zero")
	}

	if err := s.validateId(songId); err != nil {
		s.logger.Warn("GetSongVersesWithPagination: invalid songId")
		return nil, err
	}

	s.logger.Infof("GetSongVersesWithPagination: fetching verses for song ID: %d", songId)
	verses, err := s.repo.GetVersesWithPagination(songId, page, pageSize)
	if err != nil {
		s.logger.Errorf("GetSongVersesWithPagination: failed to fetch verses: %v", err)
		return nil, err
	}
	s.logger.Infof("GetSongVersesWithPagination: successfully fetched %d verses for song ID: %d", len(verses), songId)
	return verses, nil
}
