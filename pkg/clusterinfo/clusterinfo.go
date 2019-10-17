package clusterinfo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ErrInitialization is returned when client initialization fails
var ErrInitialization = errors.New("initialization")

// ErrNotFound is returned when a resource could not be found
var ErrNotFound = errors.New("not found")

// Client is used to interface with a Kubernetes cluster
type Client interface {
	Nodes() NodesClient
}

type k8sClient struct {
	config        *rest.Config
	coreService   *kubernetes.Clientset
	metricService *metricsv.Clientset
	nodesClient   NodesClient
}

func (c k8sClient) Nodes() NodesClient {
	if c.nodesClient == nil {
		c.nodesClient = k8sNodesClient{
			coreService:   c.coreService,
			metricService: c.metricService,
		}
	}
	return c.nodesClient
}

type newClientOutsideClusterOptions struct {
	configFileLocation string
}

// NewClientOutsideClusterOption can be passed when initializing a Client
// for use outside a cluster to alter default parameters
type NewClientOutsideClusterOption interface {
	apply(*newClientOutsideClusterOptions)
}

type newClientOutsideClusterOptionFunc func(*newClientOutsideClusterOptions)

func (f newClientOutsideClusterOptionFunc) apply(o *newClientOutsideClusterOptions) {
	f(o)
}

// WithConfigLocation is an option to set a custom Kubernetes config location
func WithConfigLocation(absolutePath string) NewClientOutsideClusterOption {
	return newClientOutsideClusterOptionFunc(func(o *newClientOutsideClusterOptions) {
		o.configFileLocation = absolutePath
	})
}

// NewClientInsideCluster initializes a Client for use inside of a cluster
func NewClientInsideCluster() (Client, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("%w (config): %s", ErrInitialization, err.Error())
	}
	return newk8sClientFromConfig(config)
}

// NewClientOutsideCluster initializes a Client for use outside of a cluster. Optional
// options can be passed to alter default parameters
func NewClientOutsideCluster(opts ...NewClientOutsideClusterOption) (Client, error) {
	options := newClientOutsideClusterOptions{
		configFileLocation: getDefaultConfigLocation(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	if _, err := os.Stat(options.configFileLocation); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w (config): not found", ErrInitialization)
	}

	config, err := clientcmd.BuildConfigFromFlags("", options.configFileLocation)
	if err != nil {
		return nil, fmt.Errorf("%w (config): %s", ErrInitialization, err.Error())
	}

	return newk8sClientFromConfig(config)
}

func newk8sClientFromConfig(config *rest.Config) (*k8sClient, error) {
	coreService, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("%w (service): %s", ErrInitialization, err.Error())
	}

	// TODO: Make metrics optional through configuration
	metricService, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("%w (service): %s", ErrInitialization, err.Error())
	}

	return &k8sClient{
		config:        config,
		coreService:   coreService,
		metricService: metricService,
	}, nil
}

func getDefaultConfigLocation() string {
	if home := homeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
