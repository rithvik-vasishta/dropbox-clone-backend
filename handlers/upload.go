package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/rithvik-vasishta/dropbox-clone/backend/aws"
	"github.com/rithvik-vasishta/dropbox-clone/backend/config"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"github.com/rithvik-vasishta/dropbox-clone/backend/utils"
)

func UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not provided"})
		return
	}
	saveAs := c.Query("save_file_name")
	if saveAs == "" {
		saveAs = fileHeader.Filename
	}
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read file"})
		return
	}

	shardSize := len(data) / config.NumShards
	remainder := len(data) % config.NumShards

	var primaryPaths []string
	var redundantPaths [][]string

	for i := 0; i < config.NumShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if i == config.NumShards-1 {
			end += remainder
		}
		shardData := data[start:end]
		filenamePart := fmt.Sprintf("%s.part%d", saveAs, i+1)

		primaryIdx := i % config.TotalNodes
		primaryNode := config.StorageNodes[primaryIdx]
		if !utils.IsRemoteNodeAlive(primaryNode) {
			for _, node := range config.StorageNodes {
				if utils.IsRemoteNodeAlive(node) {
					fmt.Printf("primary %s down, falling back to %s\n", primaryNode, node)
					primaryNode = node
					break
				}
			}
		}
		if err := aws.UploadShardToNode(primaryNode, filenamePart, shardData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		primaryPaths = append(primaryPaths, primaryNode+"/shards/"+filenamePart)

		var reps []string
		count := 0
		for r := 1; r <= config.TotalNodes && count < config.ReplicationFactor; r++ {
			idx := (primaryIdx + r) % config.TotalNodes
			node := config.StorageNodes[idx]
			if !utils.IsRemoteNodeAlive(node) {
				fmt.Printf("[⚠] replica %s down, skip\n", node)
				continue
			}
			if err := aws.UploadShardToNode(node, filenamePart, shardData); err != nil {
				fmt.Printf("[‼] upload to %s failed: %v\n", node, err)
				continue
			}
			reps = append(reps, node+"/shards/"+filenamePart)
			count++
		}
		if count < config.ReplicationFactor {
			fmt.Printf("[⚠] only %d/%d replicas for shard %d\n",
				count, config.ReplicationFactor, i+1)
		}
		redundantPaths = append(redundantPaths, reps)
	}

	_, err = db.DB.Exec(`
		INSERT INTO file_metadata
		  (filename, num_shards, primary_shards, redundant_shards)
		VALUES ($1,$2,$3,$4)
	`, saveAs,
		config.NumShards,
		pq.Array(primaryPaths),
		pq.Array(redundantPaths),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "uploaded with redundancy",
		"filename":         saveAs,
		"primary_shards":   primaryPaths,
		"redundant_shards": redundantPaths,
	})
}
