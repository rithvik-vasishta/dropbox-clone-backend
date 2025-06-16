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
	"strings"
)

func parse2DArray(raw string) [][]string {
	raw = strings.Trim(raw, "{}")
	rows := strings.Split(raw, "},{")

	var result [][]string
	for _, row := range rows {
		row = strings.Trim(row, "{}")
		parts := strings.Split(row, ",")
		result = append(result, parts)
	}
	return result
}

func DownloadFile(c *gin.Context) {
	filename := c.Param("file")
	fmt.Println("Requested Filename", filename)

	var primaryPaths []string
	var redundantRaw string
	err := db.DB.QueryRow(`
	SELECT primary_shards, REPLACE(redundant_shards::text, '\"', '') 
	FROM file_metadata 
	WHERE filename = $1
`, filename).Scan(
		pq.Array(&primaryPaths),
		&redundantRaw,
	)

	redundantPaths := parse2DArray(redundantRaw)

	if err != nil {
		fmt.Printf("Error querying shards: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")

	for i, primaryPath := range primaryPaths {
		var part *os.File
		part, err = os.Open(primaryPath)
		if err != nil {
			fmt.Println("[-] Primary shard", i, "failed:", primaryPath, "→", err)

			// Try replicas
			replicas := redundantPaths[i]
			replicaUsed := false
			for _, replicaPath := range replicas {
				part, err = os.Open(replicaPath)
				if err == nil {
					fmt.Println("[++] Using replica for shard", i, ":", replicaPath)
					replicaUsed = true
					break
				}
			}

			if !replicaUsed {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("All replicas failed for shard %d", i)})
				return
			}
		} else {
			fmt.Println("[+] Primary shard", i, "OK:", primaryPath)
		}

		func() {
			defer part.Close()
			_, err = io.Copy(c.Writer, part)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to stream shard %d", i)})
			}
		}()
	}

	fmt.Println("File stream complete for", filename)
}
