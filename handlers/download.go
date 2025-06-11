package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"io"
	"net/http"
	"os"
)

func DownloadFile(c *gin.Context) {
	filename := c.Param("file")
	fmt.Println("FILENAME", filename)

	var shardPaths []string
	err := db.DB.QueryRow(`SELECT shard_paths FROM file_metadata WHERE filename = $1`, filename).
		Scan(pq.Array(&shardPaths))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")

	for _, path := range shardPaths {
		part, err := os.Open(path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read shard: " + path})
			return
		}
		defer part.Close()

		_, err = io.Copy(c.Writer, part)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write shard to response"})
			return
		}
	}

}
