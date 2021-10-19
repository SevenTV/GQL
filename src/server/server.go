package server

import (
	"net"
	"time"

	"github.com/SevenTV/GQL/src/global"
	"github.com/SevenTV/GQL/src/server/middleware"
	v3 "github.com/SevenTV/GQL/src/server/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func New(gCtx global.Context) <-chan struct{} {
	ln, err := net.Listen(gCtx.Config().Http.Type, gCtx.Config().Http.URI)
	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage:        true,
		DisablePreParseMultipartForm: true,
		DisableKeepalive:             true,
		ReadTimeout:                  time.Second * 10,
		WriteTimeout:                 time.Second * 10,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Node-ID", gCtx.Config().NodeName)
		return c.Next()
	})
	app.Use(middleware.Logger())

	// v3
	v3.API(gCtx, app.Group("/v3"))

	// 404
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(&fiber.Map{
			"status":  404,
			"message": "Not Found",
		})
	})

	go func() {
		err = app.Listener(ln)
		if err != nil {
			logrus.WithError(err).Fatal("failed to start http server")
		}
	}()

	done := make(chan struct{})
	go func() {
		<-gCtx.Done()
		_ = app.Shutdown()
		close(done)
	}()

	return done
}
