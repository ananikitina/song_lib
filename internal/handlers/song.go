package handlers

import (
	"net/http"

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

func (h *SongHandler) AddSongHandler(c *gin.Context) {
	var req struct {
		Group string `json:"group" binding:"required"`
		Song  string `json:"song" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debugf("AddSongHandler: invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("AddSongHandler: adding song: %s, and group: %s", req.Song, req.Group)
	song, err := h.songService.AddSong(req.Group, req.Song)
	if err != nil {
		h.logger.Debugf("AddSongHandler: failed to add song: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("AddSongHandler: song added successfully")
	c.JSON(http.StatusOK, gin.H{"data": song})

}
