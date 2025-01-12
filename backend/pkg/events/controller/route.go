package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/spezifisch/rueder3/backend/pkg/helpers"
)

// DefaultRoute godoc
// @Summary A route that echoes the JWT claims
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} dict
// @Failure 400 {object} httputil.HTTPError
// @Failure 401 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router / [get]
func (con *Controller) DefaultRoute(c *fiber.Ctx) error {
	claims := helpers.GetFiberAuthClaims(c)
	if claims == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(fiber.Map{
		"ping":   "pong",
		"msg":    "default route of " + c.App().Config().AppName,
		"claims": claims,
	})
}

// SSE godoc
// @Summary Server-Side Events endpoint
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} dict
// @Failure 400 {object} httputil.HTTPError
// @Failure 401 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /sse [get]
func (con *Controller) SSE(c *fiber.Ctx) error {
	claims := helpers.GetFiberAuthClaims(c)
	if claims == nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	userID := claims.ID
	startTime := time.Now().UnixNano()

	// based on https://github.com/gofiber/recipes/blob/73e31998b30239a9823d6ef55c01e6eade8587cf/sse/main.go
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// NOTE do not access anything from the fiber/fasthttp context in here (only copies like userID)
		logBase := log.WithField("userID", userID).WithField("startTime", startTime)
		logBase.Info("connected")

		eventUserState, err := con.eventRepo.ConnectUser(userID)
		if err != nil {
			logBase.WithError(err).Error("couldn't connect to message queue")
			return
		}

		ticker := time.NewTicker(5 * time.Second)
		var i int
		for {
			quit := false

			select {
			case <-ticker.C:
				// we need to send something every 30s or the browser closes the connection
				i++
				payload := fmt.Sprintf("%d - the time is %v", i, time.Now())
				fmt.Fprintf(w, "event: message\ndata: Message: %s\n\n", payload)

				err := w.Flush()
				if err != nil {
					logBase.WithError(err).Info("disconnected")
					quit = true
				}
			case eventMsg := <-eventUserState.Channel:
				// send user events
				payload, err := json.Marshal(eventMsg.Payload)
				if err != nil {
					logBase.WithError(err).Info("failed re-serializing message")
				} else {
					// make sure there are no double newlines in payload so SSE doesn't break
					nlPayload := bytes.Replace(payload, []byte("\n\n"), []byte(" "), -1)
					fmt.Fprintf(w, "event: message\ndata: %s\n\n", nlPayload)

					err := w.Flush()
					if err != nil {
						logBase.WithError(err).Info("disconnected")
						quit = true
					}
				}
			}

			if quit {
				break
			}
		}

		logBase.Info("cleaning up")
		eventUserState.Close <- struct{}{}
		logBase.Info("cleaned up")
	}))

	return nil
}
