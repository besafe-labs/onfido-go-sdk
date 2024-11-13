package onfido_test

import (
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/stretchr/testify/assert"
)

var defaultRetries = onfido.WithRetries(30, 5*time.Second)

func setup(token string, opts ...onfido.ClientOption) (client *onfido.Client, teardown func(), err error) {
	client, err = onfido.NewClient(token, opts...)
	if err != nil {
		return
	}

	teardown = func() {
		client.Close()
	}

	return
}

func TestNewClient(t *testing.T) {
	t.Run("Should create NewClient without errors", func(t *testing.T) {
		client, _, err := setup("token")
		assert.NoErrorf(t, err, "error should be nil, got %v", err)
		assert.NotNil(t, client, "client should not be nil")
	})

	t.Run("Should return error when token is empty", func(t *testing.T) {
		_, _, err := setup("")
		assert.Error(t, err, "error should not be nil")
	})

	t.Run("Should set regions successfully", func(t *testing.T) {
		client, _, _ := setup("token", onfido.WithRegion(onfido.API_REGION_US))
		assert.Equal(t, "https://api.us.onfido.com/v3.6", client.Endpoint, "endpoint should be set to US region")

		client, _, _ = setup("token", onfido.WithRegion(onfido.API_REGION_CA))
		assert.Equal(t, "https://api.ca.onfido.com/v3.6", client.Endpoint, "endpoint should be set to CA region")

		client, _, _ = setup("token", onfido.WithRegion(onfido.API_REGION_EU))
		assert.Equal(t, "https://api.eu.onfido.com/v3.6", client.Endpoint, "endpoint should be set to EU region")
	})

	t.Run("Should set retries and retry wait time successfully", func(t *testing.T) {
		wait := 5 * time.Second
		client, _, _ := setup("token", onfido.WithRetries(3, wait))
		assert.Equal(t, 3, client.Retries, "retries should be set to 3")
		assert.Equal(t, wait, client.RetryWait, "retry wait time should be set to 5")
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("Should close the client without errors", func(t *testing.T) {
		_, teardown, _ := setup("token")
		teardown()
	})
}
