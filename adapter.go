package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var ch *chan map[string]interface{}
var token string
var secret string

func Boot(c *chan map[string]interface{}) {
	ch = c
	token = os.Getenv("LXBOT_KOKOROIO_ACCESSTOKEN")
	secret = os.Getenv("LXBOT_KOKOROIO_CALLBACKSECRET")

	go listen()
}

func Send(msg map[string]interface{}) {
	channelId := msg["room"].(map[string]interface{})["id"].(string)
	message := msg["message"].(map[string]interface{})["text"].(string)
	_ = send(channelId, message)
}

func Reply(msg map[string]interface{}) {

}

func send(channelId string, msg string) error {
	u := "https://kokoro.io/api/v1/bot/channels/"+channelId+"/messages"

	v := url.Values{}
	v.Set("message", msg)

	req, err := http.NewRequest("POST", u, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
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
	m := map[string]interface{}{}
	if err := c.Bind(&m); err != nil {
		return c.NoContent(406)
	}

	resErr := c.NoContent(202)

	fmt.Println(m)

	*ch <- map[string]interface{}{
		"user": map[string]interface{}{
			"id": m["profile"].(map[string]interface{})["screen_name"],
			"name": m["display_name"],
		},
		"room": map[string]interface{}{
			"id": m["channel"].(map[string]interface{})["id"],
			"name": m["channel"].(map[string]interface{})["channel_name"],
			"description": m["channel"].(map[string]interface{})["description"],
		},
		"message": map[string]interface{}{
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