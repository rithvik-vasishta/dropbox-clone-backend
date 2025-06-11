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
	var primaryShardPaths []string
	var redundantShardPaths [][]string

	for i := 0; i < numShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if i == numShards-1 {
			end += remainder
		}

		shardData := data[start:end]

		primaryNode := i % totalNodes
		primaryDir := filepath.Join("shards", fmt.Sprintf("node%d", primaryNode+1))
		os.MkdirAll(primaryDir, os.ModePerm)

		//shardDir := filepath.Join("shards", fmt.Sprintf("node%d", i+1))
		//err := os.MkdirAll(shardDir, os.ModePerm)
		//if err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shard directory"})
		//	return
		//}

		shardFilename := fmt.Sprintf("%s.part%d", saveAs, i+1)
		//if err := os.WriteFile(shardFilename, shardData, 0644); err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write shard"})
		//	return
		//}
		//
		//shardPaths = append(shardPaths, shardFilename)
		primaryPath := filepath.Join(primaryDir, shardFilename)
		if err := os.WriteFile(primaryPath, shardData, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write primary shard"})
			return
		}
		primaryShardPaths = append(primaryShardPaths, primaryPath)

		var replicas []string
		for r := 1; r <= replicationFactor; r++ {
			replicaNode := (primaryNode + r) % totalNodes
			replicaDir := filepath.Join("shards", fmt.Sprintf("node%d", replicaNode+1))
			os.MkdirAll(replicaDir, os.ModePerm)

			replicaPath := filepath.Join(replicaDir, shardFilename)
			if err := os.WriteFile(replicaPath, shardData, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write replica shard"})
				return
			}
			replicas = append(replicas, replicaPath)
		}
		redundantShardPaths = append(redundantShardPaths, replicas)
	}

	_, err = db.DB.Exec(`
        INSERT INTO file_metadata (filename, num_shards, primary_shards, redundant_shards)
        VALUES ($1, $2, $3, $4)
    `, saveAs, numShards, pq.Array(primaryShardPaths), pq.Array(redundantShardPaths))
	if err != nil {
		fmt.Println("DB Failed to insert file metadata:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store metadata in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "File uploaded and sharded with redundancy",
		"filename":         saveAs,
		"primary_shards":   primaryShardPaths,
		"redundant_shards": redundantShardPaths,
	})
}
