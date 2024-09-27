package domain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
		s.logger.Debugf("AddSong: failed to save song to database: %v", err)
		return nil, err
	}

	s.logger.Infof("AddSong: song created with ID: %d", song.ID)
	return song, nil
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
