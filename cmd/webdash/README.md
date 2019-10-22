# webdash

A simple dashboard for monitoring your Kubernetes cluster

## Giving access to the cluster
### In cluster
In order to use webdash inside your cluster, you will need to give access to listing and reading resources:
```
$ kubectl create clusterrole node-reader --verb=get,list,watch --resource=nodes,nodes/status

$ kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default

$ kubectl create clusterrolebinding default-node-reader-binding --clusterrole=node-reader --serviceaccount=default:default
```

A template Kubernetes deployment configuration can be found in `webdash.yaml`.
### Outisde cluster
If you want to run webdash outside of your cluster (this is the default), then you need to make sure you have a kube config available. By default, this is found in `~/.kube/config`.

#### Using Docker
To start webdash using Docker, do the following:
```
$ docker build webdash .

$ docker run -p 8000:8000 --rm -v "$HOME/.kube/config":"/.kube/config" -e CONN_PORT=8000 --name webdash webdash
```

You can now view the dashboard at `http://localhost:8000`.