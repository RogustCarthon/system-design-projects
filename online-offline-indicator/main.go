package main

import (
	"context"
	"online_offline_indicator/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	svc, err := service.NewService(context.Background())
	if err != nil {
		panic(err)
	}
	r.GET("/status/:userId", svc.GetStatus)
	r.POST("/status/:userId", svc.Heartbeat)
	r.Run(":8080")
}
