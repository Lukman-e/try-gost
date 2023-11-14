package connector

import (
	"testing"

	"github.com/Lukman-e/try-gost/internal/env"
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
}

func TestLoadDatabase(t *testing.T) {
	db := LoadDatabase()
	if db == nil {
		t.Error("db shouldn't nil")
	}
}

func TestLoadRedisCache(t *testing.T) {
	rds := LoadRedisCache()
	if rds == nil {
		t.Error("rds shouldn't nil")
	}
}
