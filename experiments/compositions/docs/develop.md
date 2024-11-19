# Development Guide

## New Developer getting started

A few pre-requisites:
1. [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
2. [go lang](https://go.dev/doc/install)
3. Make tooling
4. [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

Fork the https://github.com/cloud-native-compositions/compositions.git repo in github. Clone the forked repo locally and create a branch for you to work on:
```shell
export GITHUB_USER=barney-s
git clone https://github.com/${GITHUB_USER}/compositions.git
cd compositions
```

## Composition

### Building locally

Build the docker image for composition core controller:
```shell
# Build the binary locally. This is a quick check to make sure the src builds
make # or `make build`

# This needs to be set. If not make file would try to use gcr.io in the current GCP project
export IMG_REGISTRY=<container registry> # example gcr.io/team123

# build the expander docker image
make docker-build
```

### Deploying to kind cluster

To build core controller and all expanders and deploy them to a local kind cluster for manual testing:
```shell
make deploy-kind 

# kubectl now points to the kind cluster with compositions built from code and deployed.
```

### Running e2e tests

To build core controller and all expanders, deploy them to a kind cluster and then run e2e tests:
```shell
make e2e-test 
```

### jinja-expander

Note: This will be moved to expanders/ folder

```shell

# build
make build-expander-jinja2
# docker image
make docker-build-expander-jinja2

# Unit-test (deploys to a kind cluster)
make unit-test-expander-jinja2
```


### getter

Note: This will be moved to expanders/ folder

```shell

# build
make build-expander-getter
# docker image
make docker-build-expander-getter

# Unit-test (deploys to a kind cluster)
make unit-test-expander-getter
```

## Expanders

Expanders in the `expanders/` folder use common `make` targets:
* `helm-expander`
* `cel-expander`

Build the docker image for the expander grpc service.
```shell
# Build the binary locally. This is a quick check to make sure the src builds
make # or `make build`

# This needs to be set. If not make file would try to use gcr.io in the current GCP project
export IMG_REGISTRY=<container registry> # example gcr.io/team123

# build the expander docker image
make docker-build
```

We can run unit-tests locally against the docker run.
```shell
# run the service in a local docker container 
make docker-run

# in a different window run
go test

# ctrl+c the docker-run once testing is done
```

We can also run tests in a kind k8s cluster:
```shell
# this (re)creates a kind cluster with name kind-kind (default name).
# to use a different name, set the env variable: KIND_CLUSTER
# export KIND_CLUSTER=expander-test
make unit-test
```
