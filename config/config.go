package config

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	DbUrl        string
	Port         string
	JWTSecert    string
	Frontend_Url string
}
type ApiConfig struct {
	UserRepo          database.UserRepository
	ArtRepo           database.ArtRepository
	EventRepo         database.EventRepository
	EventAttendeeRepo database.EventAttendeesRepository
	ArtMetaDataRepo   database.ArtMetaDataRepository
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg = loadConfig()
	})
	return cfg
}

func getEnvLocation(path string) (string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.Name() == ".env" {
			return fmt.Sprintf("%s/.env", path), nil
		}
	}
	parent := filepath.Dir(path)
	if parent == path {
		return "", fmt.Errorf(".env not found")
	}
	return getEnvLocation(parent)

}
func loadConfig() *Config {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
	}
	envPath, err := getEnvLocation(path)
	if err != nil {
		fmt.Println(err.Error())
	}
	godotenv.Load(envPath)
	dbUrl := os.Getenv("DATABASE_URL")
	Port := os.Getenv("PORT")
	Jwt := os.Getenv("JWT_SECERT")
	fUrl := os.Getenv("FRONTEND_URL")

	return &Config{
		DbUrl:        dbUrl,
		Port:         Port,
		JWTSecert:    Jwt,
		Frontend_Url: fUrl,
	}

}
func DbQuries() (*ApiConfig, error) {
	apiConfig := &ApiConfig{}
	config := GetConfig()
	conn, err := sql.Open("postgres", config.DbUrl)
	if err != nil {
		return nil, err
	}
	query := database.New(conn)
	if query == nil {
		return nil, errors.New("Connection Failed")
	}
	apiConfig.UserRepo = query
	apiConfig.ArtRepo = query
	apiConfig.EventRepo = query
	apiConfig.EventAttendeeRepo = query
	apiConfig.ArtMetaDataRepo = query
	return apiConfig, nil
}
