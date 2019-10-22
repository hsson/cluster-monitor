# Monitoring Kubernetes

This is a small project I'm working on for interfacing with my Kubernetes cluster.

The main point of interest is the `clusterinfo` package found in `pkg/clusterinfo/`. This is a high level API client for interfacing with a Kubernetes cluster. It simplifies things a lot compared to e.g. the official `client-go` API client.

There are two implementations using the `clusterinfo` package; `litectl` and `webdash`. The `litectl` tool can be found in `cmd/litectl`, and is a liteweight implementation that resembles `kubectl`. The `webdash` can be found in `cmd/webdash`, and is a simple web dashboard that displays information about your cluster in the browser.