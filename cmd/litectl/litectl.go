package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/hsson/cluster-monitor/pkg/clusterinfo"
)

var (
	inCluster  = flag.Bool("in-cluster", false, "Specify when running this within the cluster")
	outputYaml = flag.Bool("yaml", true, "Output YAML")
	outputJSON = flag.Bool("json", false, "Output JSON")

	configLocation = flag.String("config-path", "", "Absolute path to kube config (optional)")
)

func main() {
	flag.Parse()
	checkFlags()
	if flag.NArg() == 0 {
		help()
		os.Exit(0)
	}

	cmd := flag.Args()[0]
	positionals := flag.Args()[1:]
	switch cmd {
	case "node":
		fallthrough
	case "nodes":
		nodes(positionals)
	default:
		help()
	}
}

func getClient() clusterinfo.Client {
	var client clusterinfo.Client
	var err error
	if *inCluster {
		client, err = clusterinfo.NewClientInsideCluster()
	} else {
		if *configLocation != "" {
			client, err = clusterinfo.NewClientOutsideCluster(clusterinfo.WithConfigLocation(*configLocation))
		} else {
			client, err = clusterinfo.NewClientOutsideCluster()
		}
	}
	if err != nil {
		exitErr(err)
	}
	return client
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func checkFlags() {
	for _, arg := range flag.Args() {
		if strings.HasPrefix(arg, "-") {
			exitErr(fmt.Errorf("flags must be specified before positionals"))
		}
	}
}

func output(obj interface{}) {
	if *outputJSON {
		printJSON(obj)
	} else if *outputYaml {
		printYAML(obj)
	} else {
		printYAML(obj)
	}
}

func printYAML(obj interface{}) {
	enc := yaml.NewEncoder(os.Stdout)
	if err := enc.Encode(obj); err != nil {
		exitErr(err)
	}
}

func printJSON(obj interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		exitErr(err)
	}
}
