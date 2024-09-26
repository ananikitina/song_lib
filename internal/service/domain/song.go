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
	// cоздаем новую песню
	song := &models.Song{
		GroupName: groupName,
		SongName:  songName,
	}

	// сохраняем песню в базу данных
	if err := s.repo.Add(song); err != nil {
		s.logger.Errorf("failed to add song to database: %v", err)
		return nil, err
	}

	// получаем информацию о песне из внешнего API
	apiResponse, err := s.getSongInfo(groupName, songName)
	if err != nil {
		s.logger.Errorf("failed to fetch song info: %v", err)
		return nil, err
	}

	s.logger.Infof("successfully fetched song info: %+v", apiResponse)

	return song, nil
}

func (s *songService) getSongInfo(group, song string) (*models.APIResponse, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Кодируем параметры group и song
	encodedGroup := url.QueryEscape(group)
	encodedSong := url.QueryEscape(song)

	// Формируем URL с закодированными параметрами
	url := fmt.Sprintf("%s/info?group=%s&song=%s", cfg.ExternalApi, encodedGroup, encodedSong)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		s.logger.Errorf("failed to fetch data from external API: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var apiResponse models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Errorf("failed to decode external API response: %v", err)
		return nil, err
	}
	return &apiResponse, nil
}
