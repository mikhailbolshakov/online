package database

import (
	"chats/system"
	"chats/redis"
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
	Redis    *redis.Redis
}

func Init() *Storage {
	s := &Storage{}
	s.Loc = s.setLocation()    //	location
	s.Instance = s.gormInit(0) //	gorm [mysql]
	s.Redis = redis.Init(0)    //	redis
	return s
}

func (s *Storage) setLocation() *time.Location {
	loc, err := system.Location()
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: system.LoadLocationError,
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
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: MysqlConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    MysqlConnectionProblemCode,
		})

		system.Reconnect(MysqlConnectionProblem, &attempt)

		return s.gormInit(attempt)
	}
	log.Println("Database connected")

	return db
}


