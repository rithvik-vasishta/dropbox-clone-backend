package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rithvik-vasishta/dropbox-clone/backend/handlers"
)

func RegisterRoutes(router *gin.Engine) {
	router.Static("/uploads", "./uploads")

	//router.POST("/upload", handlers.UploadFile)
	router.POST("/upload/*file", handlers.UploadFile)
	//router.GET("/download/:file", handlers.DownloadFile)
	router.GET("/download/*file", handlers.DownloadFile)
}
