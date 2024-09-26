package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ananikitina/song_lib/config"
	"github.com/ananikitina/song_lib/internal/handlers"
	"github.com/ananikitina/song_lib/internal/repository/postgresql"
	"github.com/ananikitina/song_lib/internal/service/domain"
	"github.com/ananikitina/song_lib/migrations"
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

	router.POST("/add-song", songHandler.AddSongHandler)

	err = router.Run(":" + cfg.AppPort)
	if err != nil {
		log.Fatalf("failed to start the server: %v", err)
		return
	}
	// TODO graceful shutdown
	log.Infof("Server has started on port:%s...", cfg.AppPort)
}
