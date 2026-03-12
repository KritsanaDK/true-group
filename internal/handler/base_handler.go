package handler

import (
	"net/http"
	"os"
	"tdg/internal/model"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/labstack/echo/v4"
)

type handlerFunc func(req interface{}, tracking *model.Tracking) (events.APIGatewayProxyResponse, error)

type BaseHandler struct {
	Debug     bool
	SkipCache bool
}

type Message struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
	Now     string `json:"now"`
}

func (h *BaseHandler) Health(c echo.Context) error {

	return c.JSON(http.StatusOK, Message{
		Name:    os.Getenv("NAME"),
		Version: os.Getenv("VERSION"),
		Env:     os.Getenv("ENV"),
		Now:     time.Now().Format("2006-01-02 15:04:05"),
	})
}
