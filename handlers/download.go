package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"

	"github.com/rithvik-vasishta/dropbox-clone/backend/aws"
	"github.com/rithvik-vasishta/dropbox-clone/backend/db"
	"github.com/rithvik-vasishta/dropbox-clone/backend/utils"
)

func parse2DArray(raw string) [][]string {
	raw = strings.Trim(raw, "{}")
	parts := strings.Split(raw, "},{")
	out := make([][]string, len(parts))
	for i, p := range parts {
		p = strings.Trim(p, "{}")
		if p == "" {
			out[i] = []string{}
		} else {
			out[i] = strings.Split(p, ",")
		}
	}
	return out
}

func DownloadFile(c *gin.Context) {
	filename := c.Param("file")
	fmt.Println("Download:", filename)

	var primaryPaths []string
	var redundantRaw string

	err := db.DB.QueryRow(`
	  SELECT primary_shards,
	         REPLACE(redundant_shards::text,'"','')
	    FROM file_metadata
	   WHERE filename = $1
	`, filename).Scan(pq.Array(&primaryPaths), &redundantRaw)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	redundantPaths := parse2DArray(redundantRaw)

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")

	//client := &http.Client{Timeout: 5 * time.Second}

	for i, url := range primaryPaths {
		parts := strings.SplitN(url, "/shards/", 2)
		node, path := parts[0], "/shards/"+parts[1]

		var resp *http.Response
		if utils.IsRemoteNodeAlive(node) {
			resp, err = aws.DownloadShardFromNode(node, path)
		} else {
			err = fmt.Errorf("node down")
		}

		if err != nil {
			fmt.Println("[-] primary shard", i, "failed:", err)
			found := false
			for _, repURL := range redundantPaths[i] {
				parts = strings.SplitN(repURL, "/shards/", 2)
				node, path = parts[0], "/shards/"+parts[1]
				if !utils.IsRemoteNodeAlive(node) {
					continue
				}
				resp, err = aws.DownloadShardFromNode(node, path)
				if err == nil {
					fmt.Println("[++] using replica shard", i)
					found = true
					break
				}
			}
			if !found {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("all replicas failed for shard %d", i+1),
				})
				return
			}
		}

		func() {
			defer resp.Body.Close()
			if _, err := io.Copy(c.Writer, resp.Body); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("failed streaming shard %d", i+1),
				})
			}
		}()
	}

	fmt.Println("Download complete for", filename)
}
