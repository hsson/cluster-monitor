package clusterinfo

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsV1Beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodesClient exposes access to cluster nodes
type NodesClient interface {
	List() (NodeList, error)
	Get(name string) (Node, error)
}

// Node represents a node in the cluster
type Node struct {
	Name       string        `json:"name"`
	Hostname   string        `json:"hostname"`
	InternalIP string        `json:"internal_ip"`
	Resources  NodeResources `json:"resources"`
	System     NodeSystem    `json:"system"`
}

// NodeResources is a collection of hardware resources
type NodeResources struct {
	// CPU is measured in cores
	CPU Resource `json:"cpu"`
	// Memory is measured in bytes
	Memory Resource `json:"memory"`
}

// NodeSystem describes system information for a node
type NodeSystem struct {
	Arch          string `json:"arch"`
	OS            string `json:"os"`
	KernelVersion string `json:"kernel_version"`
	OSImage       string `json:"os_image"`
}

// Resource represents an allocatable hardware resource
type Resource struct {
	Capacity int64   `json:"capacity"`
	Usage    float64 `json:"usage"`
}

// NodeList contains nodes, where each node can be indexed by its name
type NodeList map[string]Node

// All returns all nodes in the NodeList
func (nl NodeList) All() []Node {
	nodes := make([]Node, 0, len(nl))
	for _, n := range nl {
		nodes = append(nodes, n)
	}
	return nodes
}

type k8sNodesClient struct {
	coreService   *kubernetes.Clientset
	metricService *metricsv.Clientset
}

func (c k8sNodesClient) List() (NodeList, error) {
	nodesCore, err := c.coreService.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodeMap := make(NodeList)
	for _, n := range nodesCore.Items {
		nodeMap[n.Name] = constructNodeFromAPINode(n)
	}

	if c.metricService == nil {
		// Metrics are not enabled
		return nodeMap, nil
	}

	nodeMetrics, err := c.metricService.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, metric := range nodeMetrics.Items {
		node, found := nodeMap[metric.Name]
		if !found {
			// Found metric for node we don't care about
			continue
		}
		nodeMap[node.Name] = node.withMetrics(metric)
	}

	return nodeMap, nil
}

func (c k8sNodesClient) Get(name string) (Node, error) {
	nodeCorePtr, err := c.coreService.CoreV1().Nodes().Get(name, metav1.GetOptions{})
	if err != nil {
		const errorTemplate = `nodes "%s" not found`
		if err.Error() == fmt.Sprintf(errorTemplate, name) {
			return Node{}, fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return Node{}, err
	}
	if nodeCorePtr == nil {
		return Node{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	node := constructNodeFromAPINode(*nodeCorePtr)

	if c.metricService == nil {
		// Metrics are not enabled
		return node, nil
	}

	metric, err := c.metricService.MetricsV1beta1().NodeMetricses().Get(name, metav1.GetOptions{})
	if metric == nil {
		// Failed to get metric
		return node, nil
	}

	return node.withMetrics(*metric), nil
}

func constructNodeFromAPINode(n v1.Node) Node {
	node := Node{
		Name: n.Name,
		Resources: NodeResources{
			CPU: Resource{
				Capacity: n.Status.Capacity.Cpu().Value(),
			},
			Memory: Resource{
				Capacity: n.Status.Capacity.Memory().Value(),
			},
		},
		System: NodeSystem{
			Arch:          n.Status.NodeInfo.Architecture,
			KernelVersion: n.Status.NodeInfo.KernelVersion,
			OS:            n.Status.NodeInfo.OperatingSystem,
			OSImage:       n.Status.NodeInfo.OSImage,
		},
	}
	const internalIP = "InternalIP"
	const hostname = "Hostname"
	for _, a := range n.Status.Addresses {
		if a.Type == internalIP {
			node.InternalIP = a.Address
		} else if a.Type == hostname {
			node.Hostname = a.Address
		}
	}
	return node
}

func (n Node) withMetrics(metric metricsV1Beta1.NodeMetrics) Node {
	cpuCap := float64(n.Resources.CPU.Capacity * 1000)
	cpuUsage := float64(metric.Usage.Cpu().MilliValue())
	n.Resources.CPU.Usage = cpuUsage / cpuCap

	memCap := float64(n.Resources.Memory.Capacity)
	memUsage := float64(metric.Usage.Memory().Value())
	n.Resources.Memory.Usage = memUsage / memCap
	return n
}
