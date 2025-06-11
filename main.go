package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"github.com/rithvik-vasishta/dropbox-clone/backend/routes"
)

func main() {
	db.Init()
	router := gin.Default()
	routes.RegisterRoutes(router)
	err := router.Run(":6969")
	if err != nil {
		panic(err)
	}
}
