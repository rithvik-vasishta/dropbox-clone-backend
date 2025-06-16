package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/rithvik-vasishta/dropbox-clone/backend/config"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"github.com/rithvik-vasishta/dropbox-clone/backend/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

numShards := config.NumShards
replicationFactor := config.ReplicationFactor
totalNodes := config.TotalNodes

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

		if !utils.IsNodeAlive(primaryDir) {
			fmt.Printf("Primary node %s down. Looking for fallback...\n", primaryDir)
			found := false
			for j := 0; j < config.TotalNodes; j++ {
				fallbackDir := fmt.Sprintf("shards/node%d", j+1)
				if utils.IsNodeAlive(fallbackDir) {
					primaryDir = fallbackDir
					found = true
					fmt.Printf("Using %s instead.\n", primaryDir)
					break
				}
			}
			if !found {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No available node for primary shard"})
				return
			}
		}

		os.MkdirAll(primaryDir, os.ModePerm)
		shardFilename := fmt.Sprintf("%s.part%d", saveAs, i+1)
		primaryPath := filepath.Join(primaryDir, shardFilename)
		if err := os.WriteFile(primaryPath, shardData, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write primary shard"})
			return
		}
		primaryShardPaths = append(primaryShardPaths, primaryPath)

		var replicas []string
		replicaCount := 0
		for r := 1; replicaCount < config.ReplicationFactor && r <= config.TotalNodes; r++ {
			replicaNode := (primaryNode + r) % totalNodes
			replicaDir := filepath.Join("shards", fmt.Sprintf("node%d", replicaNode+1))
			if !utils.IsNodeAlive(replicaDir) {
				fmt.Printf("Replica node %s down. Skipping.\n", replicaDir)
				continue
			}
			os.MkdirAll(replicaDir, os.ModePerm)

			replicaPath := filepath.Join(replicaDir, shardFilename)
			if err := os.WriteFile(replicaPath, shardData, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write replica shard"})
				return
			}
			replicaCount++
			replicas = append(replicas, replicaPath)
		}
		if replicaCount < config.ReplicationFactor {
			fmt.Printf("Only %d replicas written for shard %d (expected %d)\n", replicaCount, i+1, config.ReplicationFactor)
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
