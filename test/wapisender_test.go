package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"id.diengs.backend/internal/config"
	"id.diengs.backend/internal/pkg"
)

func TestWapiSender_SendOne(t *testing.T) {
	viperConfig := config.NewViper()
	logger := config.NewLogger(viperConfig)

	client := pkg.NewWapiSenderClient(viperConfig, logger)

	// Ganti dengan nomor HP yang valid untuk menerima pesan test
	// targetPhone := "6285962354321"
	targetPhone := "6285962354321"

	err := client.SendOne(targetPhone, "Test pesan dari WapiSender - diengsid backend")

	assert.NoError(t, err)
}

func TestWapiSender_NormalizePhone(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"081234567890", "6281234567890"},
		{"+6281234567890", "6281234567890"},
		{"6281234567890", "6281234567890"},
		{"  081234567890 ", "6281234567890"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			result := pkg.NormalizePhone(c.input)
			assert.Equal(t, c.expected, result)
		})
	}
}
