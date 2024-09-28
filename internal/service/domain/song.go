package domain

import (
	"encoding/json"
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

type songService struct {
	repo   repository.SongRepository
	logger *logrus.Logger
}

func NewSongService(repo repository.SongRepository, logger *logrus.Logger) service.SongService {
	return &songService{
		repo:   repo,
		logger: logger,
	}
}

func (s *songService) GetSongInfo(groupName, songName string) (*models.SongDetail, error) {
	s.logger.Infof("GetSongInfo: fetching song info for group: %s and song: %s", groupName, songName)

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Кодируем параметры group и song
	encodedGroup := url.QueryEscape(groupName)
	encodedSong := url.QueryEscape(songName)

	// Формируем URL с закодированными параметрами
	url := fmt.Sprintf("%s/info?group=%s&song=%s", cfg.ExternalApi, encodedGroup, encodedSong)

	resp, err := http.Get(url)
	if err != nil {
		s.logger.Debugf("GetSongInfo: failed to fetch data from external API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Body == nil {
		s.logger.Debugf("GetSongInfo: received empty response body")
		return nil, fmt.Errorf("received empty response body")
	}

	if resp.StatusCode == http.StatusBadRequest {
		s.logger.Debugf("GetSongInfo: received 400 Bad Request from external API")
		return nil, fmt.Errorf("bad request: group or song name is invalid")
	}
	if resp.StatusCode == http.StatusInternalServerError {
		s.logger.Debugf("GetSongInfo: received 500 Internal Server Error from external API")
		return nil, fmt.Errorf("internal server error from external API")
	}
	if resp.StatusCode != http.StatusOK {
		s.logger.Debugf("GetSongInfo: unexpected status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var songDetail models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		s.logger.Debugf("GetSongInfo: failed to decode external API response: %v", err)
		return nil, fmt.Errorf("failed to decode song info response: %w", err)
	}

	s.logger.Infof("GetSongInfo: successfully fetched song info for group: %s and song: %s", groupName, songName)
	return &songDetail, nil
}

func (s *songService) AddSong(groupName, songName string) (*models.Song, error) {
	s.logger.Infof("AddSong: creating song: %s with group: %s", songName, groupName)

	// получаем информацию о песне из внешнего API
	songDetail, err := s.GetSongInfo(groupName, songName)
	if err != nil {
		s.logger.Errorf("failed to fetch song info: %v", err)
		return nil, err
	}

	// cоздаем новую песню с инфо из API
	song := &models.Song{
		GroupName:   groupName,
		SongName:    songName,
		ReleaseDate: songDetail.ReleaseDate,
		Text:        songDetail.Text,
		Link:        songDetail.Link,
	}

	// сохраняем песню в базу данных
	if err := s.repo.Add(song); err != nil {
		s.logger.Errorf("AddSong: failed to save song to database: %v", err)
		return nil, err
	}

	s.logger.Infof("AddSong: song created with ID: %d", song.ID)
	return song, nil
}

func (s *songService) GetSongById(id uint) (*models.Song, error) {
	s.logger.Infof("GetSongById: getting song with id: %d", id)
	song, err := s.repo.GetById(id)
	if err != nil {
		s.logger.Debugf("GetSongById: failed to fetch song from database: %v", err)
		return nil, err
	}

	s.logger.Infof("GetSongById: got song with ID: %d", song.ID)
	return &models.Song{
		ID:          song.ID,
		GroupName:   song.GroupName,
		SongName:    song.SongName,
		ReleaseDate: song.ReleaseDate,
		Text:        song.Text,
		Link:        song.Link,
		CreatedAt:   song.CreatedAt,
		UpdatedAt:   song.UpdatedAt,
	}, nil
}

func (s *songService) GetAllSongs() ([]models.Song, error) {
	s.logger.Info("GetAllSongs: fetching all songs")
	songs, err := s.repo.GetAll()
	if err != nil {
		s.logger.Errorf("GetAllSongs: failed to fetch songs: %v", err)
		return nil, err
	}

	return songs, nil
}

func (s *songService) UpdateSong(songId uint, updatedSong models.Song) (*models.Song, error) {
	s.logger.Infof("UpdateSong: updating song with ID: %d", songId)

	song, err := s.repo.GetById(songId)
	if err != nil {
		s.logger.Errorf("UpdateSong: failed to update song: %v", err)
		return nil, err
	}

	if updatedSong.GroupName != "" {
		song.GroupName = updatedSong.GroupName
	}
	if updatedSong.SongName != "" {
		song.SongName = updatedSong.SongName
	}
	if updatedSong.ReleaseDate != "" {
		song.ReleaseDate = updatedSong.ReleaseDate
	}
	if updatedSong.Text != "" {
		song.Text = updatedSong.Text
	}
	if updatedSong.Link != "" {
		song.Link = updatedSong.Link
	}

	song.UpdatedAt = time.Now()

	if err := s.repo.Update(song); err != nil {
		s.logger.Errorf("UpdateSong: failed to update song: %v", err)
		return nil, err
	}

	s.logger.Infof("UpdateSong: song updated with ID: %d", song.ID)
	return &models.Song{
		ID:          song.ID,
		GroupName:   song.GroupName,
		SongName:    song.SongName,
		ReleaseDate: song.ReleaseDate,
		Text:        song.Text,
		Link:        song.Link,
		CreatedAt:   song.CreatedAt,
		UpdatedAt:   time.Now(),
	}, nil
}

func (s *songService) DeleteSong(id uint) error {
	s.logger.Infof("DeleteSong: deleting song with ID: %d", id)
	if err := s.repo.Delete(id); err != nil {
		s.logger.Errorf("DeleteSong: failed to delete song: %v", err)
		return err
	}
	s.logger.Infof("DeleteSong: song deleted with ID: %d", id)
	return nil
}

func (s *songService) GetSongsWithFiltersAndPagination(filters map[string]interface{}, page int, pageSize int) ([]models.Song, error) {
	s.logger.Infof("GetSongsWithFiltersAndPagination: fetching songs with filters and pagination: filters=%d page=%d, pageSize=%d",
		len(filters), page, pageSize)
	songs, err := s.repo.GetWithFiltersAndPagination(filters, page, pageSize)
	if err != nil {
		s.logger.Debugf("GetSongsWithFiltersAndPagination: failed to fetch songs: %v", err)
		return nil, err
	}

	var allSongs []models.Song
	for _, fetchedSong := range songs {
		song := models.Song{
			ID:          fetchedSong.ID,
			GroupName:   fetchedSong.GroupName,
			SongName:    fetchedSong.SongName,
			ReleaseDate: fetchedSong.ReleaseDate,
			Text:        fetchedSong.Text,
			Link:        fetchedSong.Link,
			CreatedAt:   fetchedSong.CreatedAt,
			UpdatedAt:   fetchedSong.UpdatedAt,
		}
		allSongs = append(allSongs, song)
	}

	return allSongs, nil
}

func (s *songService) GetSongVersesWithPagination(songId uint, page int, pageSize int) ([]string, error) {
	s.logger.Infof("GetSongVersesWithPagination: fetching verses for song ID: %d", songId)
	verses, err := s.repo.GetVersesWithPagination(songId, page, pageSize)
	if err != nil {
		s.logger.Errorf("GetSongVersesWithPagination: failed to fetch verses: %v", err)
		return nil, err
	}
	s.logger.Infof("GetSongVersesWithPagination: successfully fetched %d verses for song ID: %d", len(verses), songId)
	return verses, nil
}
