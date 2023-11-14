package env

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	config    Config
	validPath = "./../../.env"
)

func TestReadConfigAndConfiguration(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()

	assert.NotEqual(t, c, config, "Expected configuration to be initialized, but it is null")
	assert.True(t, c.AppPort >= 1, "AppPort is less than 1; it should be a positive integer")
	assert.NotEmpty(t, c.DatabasePort, "DatabasePort is empty; it should have a valid value")
	assert.NotEmpty(t, c.PublicKey, "PublicKey is empty; it should have a valid value")
	assert.NotEmpty(t, c.PrivateKey, "PrivateKey is empty; it should have a valid value")
	assert.NotEmpty(t, c.RedisURI, "RedisURI is empty; it should have a valid value")
	assert.True(t, c.SMTPPort >= 1, "SMTPPort is less than 1; it should be a positive integer")
}

func TestConfigGetPublicKeyAndGetPrivateKey(t *testing.T) {
	defer func() {
		r := recover()
		assert.Nil(t, r, "should not panic")
	}()

	ReadConfig(validPath)
	c := Configuration()

	rawDbURI := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s"
	dbURI := c.GetDatabaseURI()
	assert.True(t, len(dbURI) >= len(rawDbURI), "should be more long")

	isProd := c.GetAppInProduction()
	assert.Equal(t, c.AppInProduction, isProd, "Expected AppInProduction to match the configuration, but it doesn't")

	pubKey := c.GetPublicKey()
	assert.NotNil(t, pubKey, "Public Key should not be nil")

	privKey := c.GetPrivateKey()
	assert.NotNil(t, privKey, "Private Key should not be nil")

	c.ShowConfig()
	c.setAppURL()
	_, parseURLErr := url.Parse(c.AppURL)
	assert.NoError(t, parseURLErr, "should not error while parsing url")
}
