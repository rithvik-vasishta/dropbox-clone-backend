package utils

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	MaxRetries    = 3
	RetryInterval = 300 * time.Millisecond
	Timeout       = 1 * time.Second
)

func IsNodeAlive(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	return true
}

func HeartbeatAllNodes(totalNodes int) map[string]bool {
	status := make(map[string]bool)
	for i := 1; i <= totalNodes; i++ {
		nodePath := fmt.Sprintf("shards/node%d", i)
		status[nodePath] = IsNodeAlive(nodePath)
	}
	return status
}

func IsRemoteNodeAlive(nodeURL string) bool {
	client := http.Client{Timeout: Timeout}
	for i := 0; i < MaxRetries; i++ {
		resp, err := client.Get(nodeURL + "/heartbeat")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		time.Sleep(RetryInterval)
	}
	return false
}
