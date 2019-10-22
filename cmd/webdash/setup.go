package main

import (
	"fmt"
	"os"

	"github.com/hsson/cluster-monitor/pkg/clusterinfo"
)

func init() {
	const clusterConfigKey = "CLUSTER_CONFIG"
	clusterConfigLocation := ""
	if val, found := os.LookupEnv(clusterConfigKey); found {
		clusterConfigLocation = val
	}

	var err error
	// IN_CLUSTER = inside | outside
	const clusterKey = "IN_CLUSTER"
	if val, found := os.LookupEnv(clusterKey); found {
		if val == "inside" {
			_clusterClient, err = clusterinfo.NewClientInsideCluster()
		} else if val == "outside" {
			if clusterConfigLocation != "" {
				_clusterClient, err = clusterinfo.NewClientOutsideCluster(clusterinfo.WithConfigLocation(clusterConfigLocation))
			} else {
				_clusterClient, err = clusterinfo.NewClientOutsideCluster()
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown value for $IN_CLUSTER: %s\n", val)
			os.Exit(1)
		}
	} else {
		// Default
		_clusterClient, err = clusterinfo.NewClientOutsideCluster()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v", err)
		os.Exit(1)
	}
}
