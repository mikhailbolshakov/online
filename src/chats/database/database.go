package database

import (
	"chats/models"
	"chats/redis"
	"chats/service"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
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
	loc, err := service.Location()
	if err != nil {
		service.SetError(&models.SystemError{
			Error:   err,
			Message: service.LoadLocationError,
			Code:    service.LoadLocationErrorCode,
		})
	}

	return loc
}

func (s *Storage) pinbaInit() *gopinba.Client {
	pinbaHost := os.Getenv("PINBA_HOST")
	pinba, err := gopinba.NewClient(pinbaHost)
	if err != nil {
		service.SetError(&models.SystemError{
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
		service.SetError(&models.SystemError{
			Error:   err,
			Message: MysqlConnectionProblem + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    MysqlConnectionProblemCode,
		})

		service.Reconnect(MysqlConnectionProblem, &attempt)

		return s.gormInit(attempt)
	}

	return db
}

/**
Вытаскиваем из окружения данные для полключения к базе
*/
func getDbParams() (string, string) {
	drive := os.Getenv("DB_DRIVE")

	params := fmt.Sprintf("%s:%s@%s(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PROTOCOL"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return drive, params
}
