package main

import (
	"flag"
	"fmt"
)

const helpText = `litectl can view resources in a kubernetes cluster

Available commands:
	nodes
	nodes <node name>
`

func help() {
	fmt.Println(helpText)
	flag.Usage()
}
