package app

import (
	"chats/system"
	"fmt"
	"gorm.io/gorm/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"time"
)

type Storage struct {
	Loc      *time.Location
	Instance *gorm.DB
	Redis    *Redis
}

func initStorage() *Storage {
	s := &Storage{}
	s.Loc = s.setLocation()
	s.Instance = s.gormInit(0)
	s.Redis = InitRedis(0)
	return s
}

func (s *Storage) setLocation() *time.Location {
	loc, err := Instance.GetLocation()
	if err != nil {
		Instance.ErrorHandler.SetError(&system.Error{
			Error:   err,
			Message: system.GetError(system.LoadLocationErrorCode),
			Code:    system.LoadLocationErrorCode,
		})
	}

	return loc
}

func (s *Storage) gormInit(attempt uint) *gorm.DB {

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%s host=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_HOST"),
	)

	cfg := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Info,   // Log level
				Colorful:      false,         // Disable color
			},
		),
	}

	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		Instance.ErrorHandler.SetError(&system.Error{
			Error:   err,
			Message: system.MysqlConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    system.MysqlConnectionProblemCode,
		})

		Instance.Inf.Reconnect(system.MysqlConnectionProblem, &attempt)

		return s.gormInit(attempt)
	}
	L().Debug("Database connected")

	return db
}


