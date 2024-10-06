package http

import "github.com/labstack/echo/v4"

func GetDefaultServer() *echo.Echo {
	e := echo.New()
	return e
}
