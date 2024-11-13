package onfido_test

import (
	"testing"
	"time"

	"github.com/besafe-labs/onfido-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	t.Run("NewClient", testNewClient)
	t.Run("ClientClose", testClientClose)
}

func testNewClient(t *testing.T) {
	t.Run("CreateWithoutErrors", func(t *testing.T) {
		client, _, err := setupClient("token")
		assert.NoErrorf(t, err, "error should be nil, got %v", err)
		assert.NotNil(t, client, "client should not be nil")
	})

	t.Run("ReturnErrorOnEmptyToken", func(t *testing.T) {
		_, _, err := setupClient("")
		assert.Error(t, err, "error should not be nil")
	})

	t.Run("SetRegionsSuccessfully", func(t *testing.T) {
		client, _, _ := setupClient("token", onfido.WithRegion(onfido.API_REGION_US))
		assert.Equal(t, "https://api.us.onfido.com/v3.6", client.Endpoint, "endpoint should be set to US region")

		client, _, _ = setupClient("token", onfido.WithRegion(onfido.API_REGION_CA))
		assert.Equal(t, "https://api.ca.onfido.com/v3.6", client.Endpoint, "endpoint should be set to CA region")

		client, _, _ = setupClient("token", onfido.WithRegion(onfido.API_REGION_EU))
		assert.Equal(t, "https://api.eu.onfido.com/v3.6", client.Endpoint, "endpoint should be set to EU region")
	})

	t.Run("SetRetriesAndRetryWaitSuccessfully", func(t *testing.T) {
		wait := 5 * time.Second
		client, _, _ := setupClient("token", onfido.WithRetries(3, wait))
		assert.Equal(t, 3, client.Retries, "retries should be set to 3")
		assert.Equal(t, wait, client.RetryWait, "retry wait time should be set to 5")
	})
}

func testClientClose(t *testing.T) {
	t.Run("CloseWithoutErrors", func(t *testing.T) {
		_, teardown, _ := setupClient("token")
		teardown()
	})
}
