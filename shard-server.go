package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func main() {
	r := gin.Default()

	r.GET("/heartbeat", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.POST("/upload-shard", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		os.MkdirAll("shards", os.ModePerm)
		dest := "shards/" + file.Filename
		if err := c.SaveUploadedFile(file, dest); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.StaticFS("/shards", http.Dir("shards"))

	r.Run(":8080")
}
