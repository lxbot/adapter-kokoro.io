package main

import (
	"github.com/labstack/echo/v4"
	"os"
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
	if err := c.Bind(m); err != nil {
		return c.NoContent(406)
	}

	_ = c.NoContent(201)

	println(m)

	*ch <- map[string]interface{}{
		"test": 1,
	}

	return nil
}

func authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") != secret {
			return c.NoContent(401)
		}
		return next(c)
	}
}