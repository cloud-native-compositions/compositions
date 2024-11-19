# Install

## Install Compositions

Assumes you have a kubernetes cluster ready and kubectl is setup to manage the cluster. To create a kubernetes cluster in the cloud of your choice, [see here](#kubernetes-cluster).

Install compositions

```shell
MANIFEST_URL=https://raw.githubusercontent.com/cloud-native-compositions/compositions/refs/heads/main/composition/release/manifest.yaml

kubectl apply -f ${MANIFEST_URL}
```

### Creating Context resource

Context is an optional resource that is created in each namespace we want to use Compositions in.
It is required for those compositions which use `context.spec.project` in their expanders.

Example Context object for GCP project :

```shell
export NAMESPACE=config-control    # namespace where KCC is setup
export PROJECT_ID=<GCP-PROJECT-ID> #replace with GCP project name/id

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

## Kubernetes Cluster

### GKE Cluster with KCC

Create a GKE cluster by following the [instructions here](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-zonal-cluster). Install the [Config Connector (KCC)](https://cloud.google.com/config-connector/docs/overview), by following the [installation instructions](https://cloud.google.com/config-connector/docs/how-to/install-manually). Please note we recommend [setting up KCC](https://cloud.google.com/config-connector/docs/how-to/install-manually#specify) in `config-control` namespace for samples to work with minimal changes.

The GCP examples in the [samples](../samples) use `config-control` namespace. That is the default KCC namespace for [Config Controller](https://cloud.google.com/kubernetes-engine/enterprise/config-controller/docs/overview). When manually installing KCC, if you picked a namespace other than `config-control`, please replace/use that namespace in the samples.


### ACK Cluster with ASO

TODO

### EKS Cluster with ACK
TODO

### GCP Config Connector with KCC, ASO, ACK

Assumes you have [installed gcloud](https://cloud.google.com/sdk/docs/install) and [initialized the gcloud](https://cloud.google.com/sdk/docs/initializing) to be used with your GCP project. We recommend using a separate [gcloud configuration](https://cloud.google.com/sdk/gcloud/reference/config/configurations) if you are trying out compositions in a different project. 

Enable the GCP APIs required:
```
gcloud services enable \
  krmapihosting.googleapis.com \
  container.googleapis.com  \
  cloudresourcemanager.googleapis.com \
  serviceusage.googleapis.com \
  anthos.googleapis.com
```

Create a [Config Controller (CC)](https://cloud.google.com/kubernetes-engine/enterprise/config-controller/docs/overview) instance with `--experimental-flags` set. CC clusters are regional. The list of supported regions for CC can be found [here](https://cloud.google.com/anthos-config-management/docs/how-to/config-controller-setup#create).

```
export CONFIG_CONTROLLER_NAME=compositions
export REGION=us-west2 #us-east7

#Note: Autopilot (--full-management) is not yet qualified with Kontrollers
gcloud alpha anthos config controller create ${CONFIG_CONTROLLER_NAME?} \
  --location ${REGION?} \
  --experimental-features Kontrollers 
```

Setup will take approximately 25 minutes. ConfigController includes a GKE cluster that is used for declarative management of your cloud resources. Refer to the [troubleshooting guide](https://cloud.google.com/kubernetes-engine/enterprise/config-controller/docs/troubleshoot) if you face any issues creating a Config Controller instance. At the end of the creation process, your environment is set up such that kubectl targets the Config Controller instance.

If you need to switch kubectl back to targeting this cluster at a later point, run:

```
gcloud anthos config controller get-credentials ${CONFIG_CONTROLLER_NAME?}   --location ${REGION?}
```

Give KCC permissions to manage the GCP project:

```
export SA_EMAIL="$(kubectl get ConfigConnectorContext -n config-control \
        -o jsonpath='{.items[0].spec.googleServiceAccount}' 2> /dev/null)"

gcloud projects add-iam-policy-binding "${PROJECT_ID?}" \
        --member "serviceAccount:${SA_EMAIL?}" \
        --role "roles/owner" \
        --project "${PROJECT_ID?}"
```