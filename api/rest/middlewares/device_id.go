package middlewares

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContextKey string

const DeviceIDKey ContextKey = "deviceId"

func NewDeviceId(domain string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		deviceId := c.Cookies("device_id")
		_, err := uuid.Parse(deviceId)
		if deviceId == "" || err != nil {
			deviceId = uuid.New().String()
			c.Cookie(&fiber.Cookie{
				Name:    "device_id",
				Value:   deviceId,
				Expires: time.Now().Add(24 * time.Hour * 365),
				Domain:  domain,
			})
		}
		c.Locals("deviceId", deviceId)

		c.SetUserContext(context.WithValue(c.UserContext(), DeviceIDKey, deviceId))
		return c.Next()
	}
}
