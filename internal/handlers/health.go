package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetHealth(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		status := "ok"
		dbStatus := "ok"

		sqlDB, err := db.DB()
		if err != nil {
			dbStatus = "error"
			status = "degraded"
		} else if err := sqlDB.Ping(); err != nil {
			dbStatus = "error"
			status = "degraded"
		}

		response := fiber.Map{
			"status":   status,
			"database": fiber.Map{"status": dbStatus},
		}

		if status != "ok" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(response)
		}
		return c.JSON(response)
	}
}
