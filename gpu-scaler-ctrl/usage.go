package main

import (
	"sync"
	"time"
)

var clusterUsage = map[string]GPUVal{}
var clusterUsageLock = sync.Mutex{}

type GPUVal struct {
	lastUsage float64
	lastSeen  time.Time
}

func setUsageForPod(hostname string, gpuUsage float64) {
	clusterUsageLock.Lock()
	defer clusterUsageLock.Unlock()

	clusterUsage[hostname] = GPUVal{
		lastUsage: gpuUsage,
		lastSeen:  time.Now(),
	}
}

func computeUsageAvg() float64 {
	clusterUsageLock.Lock()
	defer clusterUsageLock.Unlock()

	tenMinsOld := time.Now().Add(-10 * time.Minute)

	newClusterUsage := map[string]GPUVal{}
	var sum float64
	for hostname, val := range clusterUsage {
		if val.lastSeen.Before(tenMinsOld) {
			continue
		}

		newClusterUsage[hostname] = val
		sum += val.lastUsage
	}

	clusterUsage = newClusterUsage

	return sum / float64(len(newClusterUsage))
}
