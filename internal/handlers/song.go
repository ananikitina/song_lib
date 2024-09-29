package handlers

import (
	"net/http"
	"strconv"

	"github.com/ananikitina/song_lib/internal/models"
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

// Получение ID песни
func (h *SongHandler) parseSongID(c *gin.Context) (uint, error) {
	songID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debugf("parseSongID: invalid song ID: %v", err)
		return 0, err
	}
	return uint(songID), nil
}

// Получение параметров пагинации
func (h *SongHandler) getPaginationParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return page, pageSize
}

// @Summary Add new song
// @Description Add a new song with a group
// @Tags songs
// @Accept json
// @Produce json
// @Param request body models.AddSongRequest true "Add song request"
// @Success 201 {object} models.Song "Song added"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /add-song [post]
func (h *SongHandler) AddSongHandler(c *gin.Context) {
	var req models.AddSongRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debugf("AddSongHandler: invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("AddSongHandler: adding song: %s, and group: %s", req.Song, req.Group)
	song, err := h.songService.AddSong(c.Request.Context(), req.Group, req.Song)
	if err != nil {
		h.logger.Debugf("AddSongHandler: failed to add song: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("AddSongHandler: song added successfully")
	c.JSON(http.StatusCreated, gin.H{"data": song})
}

// @Summary Update song
// @Description Update song details by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body models.Song true "Song details to update"
// @Success 200 {object} models.Song "Song updated"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /update-song/{id} [put]
func (h *SongHandler) UpdateSongHandler(c *gin.Context) {
	songID, err := h.parseSongID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
		return
	}

	var updateReq models.Song
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		h.logger.Debugf("UpdateSongHandler: invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("UpdateSongHandler: updating song with ID: %d", songID)
	updatedSong, err := h.songService.UpdateSong(songID, updateReq)
	if err != nil {
		h.logger.Debugf("UpdateSongHandler: failed to update song: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("UpdateSongHandler: song updated with ID: %d", songID)
	c.JSON(http.StatusOK, gin.H{"data": updatedSong})
}

// @Summary Delete song
// @Description Delete song by ID
// @Tags songs
// @Produce json
// @Param id path int true "Song ID"
// @Success 200 {object} map[string]interface{} "Song deleted"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /delete-song/{id} [delete]
func (h *SongHandler) DeleteSongHandler(c *gin.Context) {
	songID, err := h.parseSongID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
		return
	}

	h.logger.Infof("DeleteSongHandler: deleting song with ID: %d", songID)
	err = h.songService.DeleteSong(songID)
	if err != nil {
		h.logger.Errorf("DeleteSongHandler: failed to delete song: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("DeleteSongHandler: song deleted with ID: %d", songID)
	c.JSON(http.StatusOK, gin.H{"message": "Song deleted successfully"})
}

// @Summary Get all songs
// @Description Retrieve a list of all songs with optional filters and pagination
// @Tags songs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Param filters query string false "Additional filters"
// @Success 200 {array} models.Song "List of songs"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /songs [get]
func (h *SongHandler) GetAllSongsHandler(c *gin.Context) {
	h.logger.Info("GetAllSongsHandler: fetching all songs")
	filters := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		filters[key] = values[0]
	}

	page, pageSize := h.getPaginationParams(c)

	songs, err := h.songService.GetSongsWithFiltersAndPagination(filters, page, pageSize)
	if err != nil {
		h.logger.Debugf("GetAllSongsHandler: failed to fetch songs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("GetAllSongsHandler: fetched %d songs", len(songs))
	c.JSON(http.StatusOK, gin.H{"songs": songs})
}

// @Summary Get song verses with pagination
// @Description Retrieve a paginated list of verses for a specific song by ID
// @Tags songs
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Success 200 {object} map[string]interface{} "Verses data"
// @Failure 400 {object} map[string]interface{} "Invalid song ID"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /songs/{id}/verses [get]
func (h *SongHandler) GetSongVersesWithPaginationHandler(c *gin.Context) {
	songID, err := h.parseSongID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID"})
		return
	}

	page, pageSize := h.getPaginationParams(c)

	h.logger.Infof("GetSongVersesWithPaginationHandler: fetching verses for song ID: %d", songID)
	verses, err := h.songService.GetSongVersesWithPagination(songID, page, pageSize)
	if err != nil {
		h.logger.Debugf("GetSongVersesWithPaginationHandler: failed to fetch verses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Infof("GetSongVersesWithPaginationHandler: fetched %d verses for song ID: %d", len(verses), songID)
	c.JSON(http.StatusOK, gin.H{"verses": verses})
}
