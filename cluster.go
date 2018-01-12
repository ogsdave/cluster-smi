package main

import (
	"github.com/patwie/cluster-smi/cluster"
	"github.com/patwie/cluster-smi/nvml"
	"os"
)

// Cluster
func FetchCluster(c *cluster.Cluster) {
	for i, _ := range c.Nodes {
		FetchNode(&c.Nodes[i])
	}
}

// Node
func InitNode(n *cluster.Node) {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	n.Name = name
	devices, _ := nvml.GetDevices()

	for i := 0; i < len(devices); i++ {
		n.Devices = append(n.Devices, cluster.Device{0, "", 0, cluster.Memory{0, 0, 0, 0}})
	}
}

func FetchNode(n *cluster.Node) {

	devices, _ := nvml.GetDevices()

	for idx, device := range devices {
		meminfo, _ := device.GetMemoryInfo()
		gpuPercent, _, _ := device.GetUtilization()
		memPercent := int(meminfo.Used / meminfo.Total)
		n.Devices[idx].Id = idx
		n.Devices[idx].Name = device.DeviceName
		n.Devices[idx].Utilization = gpuPercent
		n.Devices[idx].MemoryUtilization = cluster.Memory{meminfo.Used, meminfo.Free, meminfo.Total, memPercent}

	}
}