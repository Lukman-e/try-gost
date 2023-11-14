package helper

import (
	"net"
	"strings"
	"testing"

	"github.com/Lukman-e/try-gost/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for i := 0; i < 25; i++ {
		s := RandomString(uint(i))
		assert.Len(t, s, i, "length of string should equal")
	}
}

func TestRandomEmails(t *testing.T) {
	for i := 1; i <= 20; i++ {
		emails := RandomEmails(uint(i))
		assert.Len(t, emails, i, "total of emails should equal")
		for _, email := range emails {
			assert.GreaterOrEqual(t, len(email), 10, "length of an email should not less than 10")
			assert.Equal(t, email, strings.ToLower(email), "email should be lowercase")
		}
	}
}

func TestRandomEmail(t *testing.T) {
	for i := 0; i < 25; i++ {
		email := RandomEmail()
		assert.NoError(t, ValidateEmails(email), "should not error")
		assert.GreaterOrEqual(t, len(email), 25, "should be more than 25")
	}
}

func TestRandomIPAddress(t *testing.T) {
	for i := 0; i < 20; i++ {
		ipRand := RandomIPAddress()
		ip := net.ParseIP(ipRand)
		assert.NotNil(t, ip, constants.ShouldNotNil)
	}
}

func TestValidateEmails(t *testing.T) {
	err1 := ValidateEmails("f", "a")
	assert.Error(t, err1, constants.ShouldErr)

	err2 := ValidateEmails("validemail098@gmail.com")
	assert.NoError(t, err2, "should not error")

	err3 := ValidateEmails("validemail0911@gmail.com", "invalidemail0987@.gmail.com")
	assert.Error(t, err3, constants.ShouldErr)

	err4 := ValidateEmails("validemail0987@gmail.com", "valid_email0987@gmail.com", "invalidemail0987@gmail.com.")
	assert.Error(t, err4, constants.ShouldErr)
}

func TestNewFiberCtx(t *testing.T) {
	c := NewFiberCtx()
	assert.NotNil(t, c, constants.ShouldNotNil)
}

func TestToTitle(t *testing.T) {
	payloads := []struct {
		str     string
		isEqual bool
	}{
		{"", true},
		{"Bca", true},
		{"A B C", true},
		{"Your Name", true},
		{"ABC", false},
		{"aa", false},
		{" bb", false},
		{"   ccc", false},
		{"a ab cc", false},
		{"your -Name", false},
		{"-m'rning", false},
	}

	for _, payload := range payloads {
		res := ToTitle(payload.str)

		if res != payload.str && payload.isEqual {
			assert.Failf(t, "should equal, but not at: %s got %s", payload.str, res)
		} else if res == payload.str && !payload.isEqual {
			assert.Failf(t, "should not equal, but equal at: %s got %s", payload.str, res)
		}
	}
}
