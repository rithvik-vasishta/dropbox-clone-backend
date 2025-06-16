package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
