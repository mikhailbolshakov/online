package infrastructure

import (
	"github.com/joho/godotenv"
	"os"
)

type Config map[string]string

/**
	Точка входа
 */
func SetEnvironment() {
	config := &Config{}
	config.read()
}

/**
	Читаем файл с конфигурациями
 */
func (config *Config) read() {
	if len(os.Args) > 1 {
		_ = godotenv.Load(os.Args[1])
	}
}