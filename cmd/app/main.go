// main.go

// @title Song Library API
// @version 1.0
// @description This is a simple API for managing songs in a song library.
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.url http://example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ananikitina/song_lib/config"
	_ "github.com/ananikitina/song_lib/docs"
	"github.com/ananikitina/song_lib/internal/handlers"
	"github.com/ananikitina/song_lib/internal/repository/postgresql"
	"github.com/ananikitina/song_lib/internal/service/domain"
	"github.com/ananikitina/song_lib/migrations"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	log := logrus.New()
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := cfg.DSN()

	var db *gorm.DB

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Errorf("failed to connect to database: %v", err)
	}

	if err := migrations.RunMigration(dsn, log); err != nil {
		log.Errorf("failed to run migrations: %v", err)
	}

	songRepository := postgresql.NewSongRepository(db, log)
	songService := domain.NewSongService(songRepository, log)
	songHandler := handlers.NewSongHandler(songService, log)

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// API routes
	// @Router /add-song [post]
	router.POST("/add-song", songHandler.AddSongHandler)
	// @Router /update-song/{id} [put]
	router.PUT("/update-song/:id", songHandler.UpdateSongHandler)
	// @Router /delete-song/{id} [delete]
	router.DELETE("/delete-song/:id", songHandler.DeleteSongHandler)
	// @Router /songs [get]
	router.GET("/songs", songHandler.GetAllSongsHandler)
	// @Router /songs/{id}/verses [get]
	router.GET("/songs/:id/verses", songHandler.GetSongVersesWithPaginationHandler)

	err = router.Run(":" + cfg.AppPort)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
		return
	}

	log.Infof("Server has started on port:%s...", cfg.AppPort)
}
