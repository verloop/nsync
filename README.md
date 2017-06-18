# Namespace Sync

Sync selected secrets from current namespace to selected namespaces.


# Installing

```bash
go get -u github.com/verloop/nSync
```

## Local

```bash
nSync --kubeconfig=$HOME/.kube/config
```

## Inside the cluster

```bash
kubectl run verloop-nsync --image=verloopio/nsync:stable
```

# Usage

## Start syncing a namespace

```bash
kubectl create ns hello
kubectl annotate namespace/hello verloop.io/managed="true" --overwrite
```

## Start syncing a secret
```bash
kubectl create secret generic my-secret --from-literal=super=secret
kubectl annotate secrets/my-secret verloop.io/managed="true" --overwrite
```

## Start syncing a configmap
```bash
kubectl create configmap global-config --from-literal=hello=world
kubectl annotate configmap/global-config verloop.io/managed="true" --overwrite
```



# Building

```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o deploy/files/nSync .
docker build -t verloopio/nsync .
docker push verloopio/nsync
```


# FAQ

## Why?
We run on a 1 namespace per customer model.

I needed something to copy over common secrets and configmaps.
