package database

import (
	"chats/redis"
	"chats/sentry"
	"chats/infrastructure"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/mkevac/gopinba"
	"log"
	"os"
	"strconv"
	"time"
)

type Storage struct {
	Loc      *time.Location
	Instance *gorm.DB
	Redis    *redis.Redis
	Pinba    *gopinba.Client
}

func Init() *Storage {
	s := &Storage{}
	s.Loc = s.setLocation()    //	location
	s.Instance = s.gormInit(0) //	gorm [mysql]
	s.Redis = redis.Init(0)    //	redis
	s.Pinba = s.pinbaInit()    //	pinba

	return s
}

func (s *Storage) setLocation() *time.Location {
	loc, err := infrastructure.Location()
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: infrastructure.LoadLocationError,
			Code:    infrastructure.LoadLocationErrorCode,
		})
	}

	return loc
}

func (s *Storage) pinbaInit() *gopinba.Client {
	pinbaHost := os.Getenv("PINBA_HOST")
	pinba, err := gopinba.NewClient(pinbaHost)
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: PinbaConnectionProblem,
			Code:    PinbaConnectionProblemCode,
			Data:    []byte(pinbaHost),
		})

		log.Println("pinba connection problem") //	TODO

		pinba = nil
	}

	return pinba
}

func (s *Storage) gormInit(attempt uint) *gorm.DB {
	db, err := gorm.Open(getDbParams())
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: MysqlConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    MysqlConnectionProblemCode,
		})

		infrastructure.Reconnect(MysqlConnectionProblem, &attempt)

		return s.gormInit(attempt)
	}
	log.Println("Database connected")

	return db
}

/**
Вытаскиваем из окружения данные для полключения к базе
*/
func getDbParams() (string, string)  {
	drive := os.Getenv("DB_DRIVE")

	params := fmt.Sprintf("user=%s password=%s dbname=%s port=%s host=%s sslmode=disable TimeZone=Europe/Moscow",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_HOST"),
	)

	return drive, params
}
