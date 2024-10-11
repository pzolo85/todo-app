package mail

import (
	"github.com/labstack/echo/v4"
	"github.com/pzolo85/todo-app/back/internal/notify"
)

type Notify interface {
	Notify(data []byte, id int, topics ...notify.Topic) error
	Subscribe(c echo.Context, topics []string) error
}
