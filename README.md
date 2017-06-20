# NSync

Sync selected secrets and configmaps across namespaces.


# How it works
NSync keeps an eye on all secrets and configmaps in the current namespace that have the annotaion `nsync.verloop.io/managed="true"`

Whenever it encounters a namespace with the annotation `nsync.verloop.io/managed="true"`, it starts syncing the secrets and configmaps to the namespace.

# Installing

```bash
go get -u github.com/verloop/nsync
```

## Local

```bash
glide install
nsync --kubeconfig=$HOME/.kube/config
```

## Inside the cluster

```bash
kubectl run verloop-nsync --image=verloopio/nsync:0.0.6
```

# Usage

## Add a namespace to be synced

```bash
kubectl create ns hello
kubectl annotate namespace/hello nsync.verloop.io/managed="true" --overwrite
```

## Start syncing a secret
```bash
kubectl create secret generic my-secret --from-literal=super=secret
kubectl annotate secrets/my-secret nsync.verloop.io/managed="true" --overwrite
```

## Start syncing a configmap
```bash
kubectl create configmap global-config --from-literal=hello=world
kubectl annotate configmap/global-config nsync.verloop.io/managed="true" --overwrite
```


# Building

```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o deploy/files/nsync .
docker build -t verloopio/nsync .
docker push verloopio/nsync
```


# FAQ

## Why make this project?
We run on a 1 namespace per customer model.

I needed something to copy over common secrets and configmaps.

## Why the name?
Because it's gonna be me.