package handlers

import (
	"github.com/ananikitina/song_lib/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SongHandler struct {
	songService service.SongService
	logger      *logrus.Logger
}

func NewSongHandler(songService service.SongService, logger *logrus.Logger) *SongHandler {
	return &SongHandler{
		songService: songService,
		logger:      logger,
	}
}

func (h *SongHandler) AddSong(c *gin.Context) {

}
