package test

import (
	"id.diengs.backend/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var app *fiber.App
var db *gorm.DB
var log *logrus.Logger

func init() {
	viperConfig := config.NewViper()
	log = config.NewLogger(viperConfig)
	db = config.NewDatabase(viperConfig, log)
	app = config.NewFiber(viperConfig)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: config.NewValidator(),
		Config:   viperConfig,
		Mail:     nil,
	})
}
