package migrations

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	migratePg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func RunMigration(dsn string, log *logrus.Logger) error {
	// Соединение с базой данных
	dbMigration, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Debugf("Failed to connect to database: %v", err)
		return err
	}
	defer func() {
		if closeErr := dbMigration.Close(); closeErr != nil {
			log.Warnf("Failed to close database connection: %v", closeErr)
		}
	}()

	// Инициализация драйвера
	driver, err := migratePg.WithInstance(dbMigration, &migratePg.Config{})
	if err != nil {
		log.Errorf("Failed to create migration driver: %v", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver)
	if err != nil {
		log.Errorf("Failed to create new migration instance: %v", err)
		return err
	}

	// Запуск миграций
	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No new migrations to apply")
			return nil
		}
		log.Debugf("Failed to run migration: %v", err)
		return err
	}

	log.Info("Migrations ran successfully")
	return nil
}
