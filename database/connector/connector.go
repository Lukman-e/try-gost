package connector

import (
	"log"
	"sync"
	"time"

	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/go-redis/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	gormDatabase     *gorm.DB
	gormDatabaseOnce sync.Once

	redisDatastore     *redis.Client
	redisDatastoreOnce sync.Once
)

// LoadDatabase func read env intenal package and
// give database connection using gorm package,
// also pings the DB before the function end.
func LoadDatabase() *gorm.DB {
	gormDatabaseOnce.Do(func() {
		// try to read env
		env.ReadConfig("./.env")
		config := env.Configuration()
		dsn := config.GetDatabaseURI()

		// try to connect to database
		var conErr error
		gormDatabase, conErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if conErr != nil {
			panic("panic while try to connect : " + conErr.Error())
		}
		if gormDatabase == nil {
			panic("error : database is nil")
		}

		// try to ping database
		database, sqlErr := gormDatabase.DB()
		if sqlErr != nil {
			log.Panicf("can't get sql-db : %s", sqlErr)
		}
		if database == nil {
			log.Panicf("can't get sql-db : database is nil")
		}
		pingErr := database.Ping()
		if pingErr != nil {
			log.Panicf("can't ping sql-db : %s", pingErr)
		}

		// config for small-to-medium web applications
		// read https://www.alexedwards.net/blog/configuring-sqldb
		database.SetMaxOpenConns(25)
		database.SetMaxIdleConns(25)
		database.SetConnMaxLifetime(time.Hour)
	})

	return gormDatabase
}

// LoadRedisCache func read env intenal package and
// give redis connection using redis external package,
// also pings the DB before the function end.
func LoadRedisCache() *redis.Client {
	redisDatastoreOnce.Do(func() {
		env.ReadConfig("./.env")
		config := env.Configuration()
		opt, err := redis.ParseURL(config.RedisURI)
		if err != nil {
			log.Panicf("can't connect to redis %s", err)
		}

		redisDatastore = redis.NewClient(opt)

		_, pingErr := redisDatastore.Ping().Result()
		if pingErr != nil {
			log.Panicf("can't ping to redis %T: %s", config.RedisURI, pingErr)
		}
	})

	return redisDatastore
}
