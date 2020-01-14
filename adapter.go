package main

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type M = map[string]interface{}

var ch *chan M
var token string
var secret string

func Boot(c *chan M) {
	ch = c
	token = os.Getenv("LXBOT_KOKOROIO_ACCESSTOKEN")
	secret = os.Getenv("LXBOT_KOKOROIO_CALLBACKSECRET")

	go listen()
}

func Send(msg M) {
	channelId := msg["room"].(M)["id"].(string)
	message := msg["message"].(M)["text"].(string)
	_ = send(channelId, message)
}

func Reply(msg M) {
	channelId := msg["room"].(M)["id"].(string)
	user := msg["user"].(M)["id"].(string)
	message := "@" + user + " " + msg["message"].(M)["text"].(string)
	_ = send(channelId, message)
}

func send(channelId string, msg string) error {
	u := "https://kokoro.io/api/v1/bot/channels/"+channelId+"/messages"

	v := url.Values{}
	v.Set("message", msg)

	req, err := http.NewRequest("POST", u, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("X-Access-Token", token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

func listen() {
	e := echo.New()
	e.GET("/", get)
	e.POST("/", post, authorize)
	_ = e.Start("0.0.0.0:1323")
}

func get(c echo.Context) error {
	return c.NoContent(200)
}

func post(c echo.Context) error {
	m := M{}
	if err := c.Bind(&m); err != nil {
		return c.NoContent(406)
	}

	resErr := c.NoContent(202)

	*ch <- M{
		"user": M{
			"id": m["profile"].(M)["screen_name"],
			"name": m["display_name"],
		},
		"room": M{
			"id": m["channel"].(M)["id"],
			"name": m["channel"].(M)["channel_name"],
			"description": m["channel"].(M)["description"],
		},
		"message": M{
			"id": m["id"],
			"text": m["plaintext_content"],
			"attachments": nil,
		},
		"raw": m,
	}

	return resErr
}

func authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") != secret {
			return c.NoContent(401)
		}
		return next(c)
	}
}