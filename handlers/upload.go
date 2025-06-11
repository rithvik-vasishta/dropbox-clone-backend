package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const numShards = 5
const replicationFactor = 2
const totalNodes = 8

func UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}

	saveAs := c.Query("save_file_name")
	if saveAs == "" {
		saveAs = fileHeader.Filename
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open uploaded file"})
		return
	}
	defer srcFile.Close()

	data, err := io.ReadAll(srcFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	shardSize := len(data) / numShards
	remainder := len(data) % numShards
	var shardPaths []string

	for i := 0; i < numShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if i == numShards-1 {
			end += remainder
		}

		shardData := data[start:end]
		shardDir := filepath.Join("shards", fmt.Sprintf("node%d", i+1))
		err := os.MkdirAll(shardDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shard directory"})
			return
		}

		shardFilename := filepath.Join(shardDir, saveAs+fmt.Sprintf(".part%d", i+1))
		if err := os.WriteFile(shardFilename, shardData, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write shard"})
			return
		}

		shardPaths = append(shardPaths, shardFilename)
	}

	_, err = db.DB.Exec(`
        INSERT INTO file_metadata (filename, num_shards, shard_paths)
        VALUES ($1, $2, $3)
    `, saveAs, numShards, pq.Array(shardPaths))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store metadata in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File sharded and uploaded successfully",
		"filename": saveAs,
		"shards":   shardPaths,
	})
}
