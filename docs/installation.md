# Install

## Kubernetes Cluster

### GKE Cluster with KCC

1. [Create a GKE cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-zonal-cluster)
2. [Instructions](https://cloud.google.com/config-connector/docs/how-to/install-manually) to install KCC on the cluster
   1. [Install KCC](https://cloud.google.com/config-connector/docs/how-to/install-manually#installing_the_operator)  
   2. [Create a SA with editor permissions](https://cloud.google.com/config-connector/docs/how-to/install-manually#addon-configuring)  
   3. [Associate the SA](https://cloud.google.com/config-connector/docs/how-to/install-manually#identity) with KCC  
   4. [Create a namespace](https://cloud.google.com/config-connector/docs/how-to/install-manually#specify) for KCC objects. We recommend using `config-control` as the namespace for samples to work with minimal changes.
3. Setup `kubectl` to target the cluster

### ACK Cluster with ASO

TODO

### EKS Cluster with ACK
TODO

### GCP Config Connector with KCC, ASO, ACK
TODO


## Install Compositions

Once the kubernetes cluster is ready, install Compositions. This step can be skipped for Config Controller cluster. 

```
# TODO Change to CNC repo path once code is merged in
MANIFEST_URL=https://raw.githubusercontent.com/GoogleCloudPlatform/k8s-config-connector/master/experiments/compositions/composition/release/manifest.yaml

kubectl apply -f ${MANIFEST_URL}
```

### Context object

Context is an optional config object that is created in each namespace we want to use Compositions in.
It is required for those compositions which use `context.spec.project` in their expanders.
For example if KCC was setup in `config-control` namespace:

```
export NAMESPACE=<where composition is to be used>
export PROJECT_ID=<replace with GCP project name/id>

kubectl apply -f - <<EOF
apiVersion: composition.google.com/v1alpha1
kind: Context
metadata:
  name: context
  namespace: ${NAMESPACE?}
spec:
  project: ${PROJECT_ID?}
EOF
```

## Sample Namespaces

Some of the GCP examples use `config-control` namespace. That is the default KCC namespace for Config Controller. When manually installing KCC, you may have picked a different namespace. Please replace use that namespace instead of `config-control`.