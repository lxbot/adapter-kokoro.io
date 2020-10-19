package main

import (
	"github.com/labstack/echo/v4"
	"github.com/lxbot/lxlib"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type M = map[string]interface{}

var ch *chan M
var token string
var secret string
var startTime time.Time

func Boot(c *chan M) {
	ch = c
	token = os.Getenv("LXBOT_KOKOROIO_ACCESSTOKEN")
	secret = os.Getenv("LXBOT_KOKOROIO_CALLBACKSECRET")
	startTime = time.Now()

	go listen()
}

func Send(msg M) {
	m, err := lxlib.NewLXMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}
	channelId := m.Room.ID
	message := m.Message.Text
	_ = send(channelId, message)
}

func Reply(msg M) {
	m, err := lxlib.NewLXMessage(msg)
	if err != nil {
		log.Println(err)
		return
	}
	channelId := m.Room.ID
	user := m.User.ID
	message := "@" + user + " " + m.Message.Text
	_ = send(channelId, message)
}

func send(channelId string, msg string) error {
	u := "https://kokoro.io/api/v1/bot/channels/" + channelId + "/messages"

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
	ms :=  new(runtime.MemStats)
	runtime.ReadMemStats(ms)

	return c.HTML(http.StatusOK, `
<!doctype html>
<html>
<head>
<title>lxbot - adapter-kokoro.io</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/water.css@2/out/water.min.css">
</head>
<body>
<h1><a href="https://lxbot.io">lxbot</a> - <a href="https://github.com/lxbot/adapter-kokoro.io">adapter-kokoro.io</a></h1>
<ul>
<li>heap: ` + strconv.FormatUint(ms.HeapAlloc / 1024, 10) + ` / `+ strconv.FormatUint(ms.TotalAlloc / 1024, 10) + ` KB</li>
<li>sys: ` + strconv.FormatUint(ms.Sys / 1024, 10) + ` KB</li>
<li>goroutine: ` + strconv.Itoa(runtime.NumGoroutine()) + `</li>
<li>uptime: ` + time.Since(startTime).String() + `</li>
</ul>
</body>
</html>
`)
}

func post(c echo.Context) error {
	m := M{}
	if err := c.Bind(&m); err != nil {
		return c.NoContent(406)
	}

	resErr := c.NoContent(202)

	*ch <- M{
		"user": M{
			"id":   m["profile"].(M)["screen_name"],
			"name": m["display_name"],
		},
		"room": M{
			"id":          m["channel"].(M)["id"],
			"name":        m["channel"].(M)["channel_name"],
			"description": m["channel"].(M)["description"],
		},
		"message": M{
			"id":          strconv.Itoa(int(m["id"].(float64))),
			"text":        m["plaintext_content"],
			"attachments": nil,
		},
		"raw": m,
	}

	return resErr
}

func authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") != secret {
			log.Println("invalid webhook secret")
			return c.NoContent(401)
		}
		return next(c)
	}
}
