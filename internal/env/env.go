package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppName           string        `env:"APP_NAME"`
	AppInProduction   bool          `env:"APP_IN_PRODUCTION"`
	AppAccessTokenTTL time.Duration `env:"APP_ACCESS_TOKEN_TTL"`
	AppPort           int           `env:"APP_PORT"`
	AppTimeZone       string        `env:"APP_TIME_ZONE"`
	AppURL            string

	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseUser     string `env:"DB_USERNAME"`
	DatabasePassword string `env:"DB_PASSWORD"`
	DatabaseName     string `env:"DB_DATABASE"`
	DatabaseURI      string

	RedisURI string `env:"REDIS_URI"`

	PublicKey  string `env:"PUBLIC_KEY"`
	PrivateKey string `env:"PRIVATE_KEY"`

	SMTPServer   string `env:"SMTP_SERVER"`
	SMTPPort     int    `env:"SMTP_PORT"`
	SMTPEmail    string `env:"SMTP_EMAIL"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	ClientURL    string `env:"CLIENT_URL"`

	BucketName  string `env:"SUPABASE_BUCKET_NAME"`
	BucketURL   string `env:"SUPABASE_URL"`
	BucketToken string `env:"SUPABASE_TOKEN"`
}

var (
	PublicKey  *[]byte
	PrivateKey *[]byte

	envFile *string
	cfg     Config

	PublicKeyReadOne  sync.Once
	PrivateKeyReadOne sync.Once
	cfgOnce           sync.Once

	paths = []string{"", "./..", "./../..", "./../../.."}
)

// ReadConfig func check .env file and read the value.
func ReadConfig(filePath string) *Config {
	cfgOnce.Do(func() {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Panicf(`.env file isn't exist/found at: "%s": %s`, filePath, err.Error())
		}
		envFile = &filePath
		log.Printf(`Reading config file: "%s"`, *envFile)
		err := cleanenv.ReadConfig(filePath, &cfg)
		if err != nil {
			err := cleanenv.ReadEnv(&cfg)
			if err != nil {
				log.Fatalf("Config error %s", err.Error())
			}
		}
	})
	return &cfg
}

// Configuration func set and update .env file
// and returning config it self.
func Configuration() Config {
	if envFile == nil {
		log.Panic(`configuration file is not set. Call ReadConfig("path_to_file") first`)
	}
	err := cleanenv.UpdateEnv(&cfg)
	if err != nil {
		log.Fatalf("Config error %s", err.Error())
	}
	cfg.setAppURL()
	return cfg
}

// GetDatabaseURI func return string URI of database.
func (c *Config) GetDatabaseURI() string {
	c.DatabaseURI = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		c.DatabaseHost, c.DatabaseUser, c.DatabasePassword, c.DatabaseName, c.DatabasePort, c.AppTimeZone,
	)

	return c.DatabaseURI
}

// GetAppInProduction func return application condition is
// under development or production ready.
func (c *Config) GetAppInProduction() bool {
	return c.AppInProduction
}

// setAppURL func combine set AppURL with localhost and port.
func (c *Config) setAppURL() {
	localAddr := fmt.Sprintf("http://127.0.0.1:%d/", c.AppPort)
	c.AppURL = localAddr
}

// GetPublicKey func gets PublicKey values.
func (c *Config) GetPublicKey() []byte {
	PublicKeyReadOne.Do(func() {
		var foundPath string
		for _, path := range paths {
			keyPath := filepath.Join(path, c.PublicKey)
			if _, err := os.Stat(keyPath); err == nil {
				foundPath = keyPath
				break
			}
		}
		if foundPath == "" {
			log.Panicf(`Publickey file isn't exist/found: "%s"`, c.PublicKey)
		}
		signKey, err := os.ReadFile(foundPath)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		PublicKey = &signKey
	})
	return *PublicKey
}

// GetPrivateKey func gets PrivateKey values.
func (c *Config) GetPrivateKey() []byte {
	PrivateKeyReadOne.Do(func() {
		var foundPath string
		for _, path := range paths {
			keyPath := filepath.Join(path, c.PrivateKey)
			if _, err := os.Stat(keyPath); err == nil {
				foundPath = keyPath
				break
			}
		}
		if foundPath == "" {
			log.Panicf(`Privatekey file isn't exist/found: "%s"`, c.PrivateKey)
		}
		signKey, err := os.ReadFile(foundPath)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		PrivateKey = &signKey
	})
	return *PrivateKey
}

// ShowConfig prints all fields that Config struct has
func (c *Config) ShowConfig() {
	fmt.Printf("%-21s: %s\n", "AppName", c.AppName)
	fmt.Printf("%-21s: %v\n", "AppInProduction", c.AppInProduction)
	fmt.Printf("%-21s: %s\n", "AppAccessTokenTTL", c.AppAccessTokenTTL)
	fmt.Printf("%-21s: %d\n", "AppPort", c.AppPort)
	fmt.Printf("%-21s: %s\n", "AppTimeZone", c.AppTimeZone)
	fmt.Printf("%-21s: %s\n", "AppURL", c.AppURL)

	fmt.Printf("%-21s: %s\n", "DatabaseHost", c.DatabaseHost)
	fmt.Printf("%-21s: %s\n", "DatabasePort", c.DatabasePort)
	fmt.Printf("%-21s: %s\n", "DatabaseUser", c.DatabaseUser)
	fmt.Printf("%-21s: %s\n", "DatabasePassword", c.DatabasePassword)
	fmt.Printf("%-21s: %s\n", "DatabaseName", c.DatabaseName)
	fmt.Printf("%-21s: %s\n", "DatabaseURI", c.DatabaseURI)

	fmt.Printf("%-21s: %s\n", "RedisURI", c.RedisURI)

	fmt.Printf("%-21s: %s\n", "PublicKey", c.PublicKey)
	fmt.Printf("%-21s: %s\n", "PrivateKey", c.PrivateKey)

	fmt.Printf("%-21s: %s\n", "SMTPServer", c.SMTPServer)
	fmt.Printf("%-21s: %d\n", "SMTPPort", c.SMTPPort)
	fmt.Printf("%-21s: %s\n", "SMTPEmail", c.SMTPEmail)
	fmt.Printf("%-21s: %s\n", "SMTPPassword", c.SMTPPassword)
	fmt.Printf("%-21s: %s\n", "ClientURL", c.ClientURL)

	// ...
	// add more
}
