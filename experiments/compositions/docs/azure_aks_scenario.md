# Scenario: AKS cluster creation & attach

This scenario will demonstrate the multi-cloud capabilities of compositions by leveraging ASO to manage Azure resources. We use the example of creating an Azure AKS cluster and attaching it to GKE Enterprise.

## Deploy an attached AKS cluster with composition

With composition, users can separate the infrastructure creation tasks to platform team and user teams. The platform team can define the composition with all the dependencies of the attached AKS cluster. On the other hand, the user team can provision the attached AKS cluster based on their needs with the composition defined by the platform team.

Here is an example of an attached AKS cluster. (The complete sample is uploaded in [github](https://github.com/cloud-native-compositions/compositions/samples/AttachedAKS). We will go through it in this guide.)

### Platform Team

#### **Step 1: Create the composition for the attached AKS cluster**

The platform administrator would create the composition in our sample:

```shell
kubectl apply -f \
https://raw.githubusercontent.com/cloud-native-compositions/compositions/main/samples/AttachedAKS/01-composition.yaml
```

This file contains a CRD, which is an interface for the composition. The user team needs to input the variables defined in this CRD for the AKS cluster.

The file also contains the composition. The composition is a combination of three resources:

* an Azure resource group  
* an AKS cluster  
* a GCP attached cluster 

These are the building blocks for  attached AKS clusters. 

#### **Step 2: Create a separate identity for the user team**

The platform team admin can create an identity for the user team, so that it can separate the identity and permissions for different teams. A summary of the steps are:

1. Create a new GCP service account for the user team  
2. Create a new Azure managed identity for the user team  
3. Create a new kubernetes namespace for the user team   
4. Assign the GCP and Azure identity to the namespace

After these steps, when the user team creates resources in this kubernetes namespace, the resources will be created with the new designated identity.

Here are the details for how to create the separate identity:

1. Create a new GCP service account for the user team:

```shell
TEAM_NAME=<a name for the user team> # example: team-aks
export NAMESPACE=${TEAM_NAME?} # you can choose a different name
TEAM_GCP_SA_NAME="${TEAM_NAME?}" # you can choose a different name
export PROJECT_ID=<GCP project for this team> 
export TEAM_GSA_EMAIL="${TEAM_GCP_SA_NAME?}@${PROJECT_ID?}.iam.gserviceaccount.com"

gcloud iam service-accounts create ${TEAM_GCP_SA_NAME?} --project ${PROJECT_ID?}

# grant this GCP service account the permission to manage resource in project
gcloud projects add-iam-policy-binding ${PROJECT_ID?} \
    --member="serviceAccount:${TEAM_GSA_EMAIL?}" \
    --role="roles/owner"

# grant KCC permission to use this service account
gcloud iam service-accounts add-iam-policy-binding \
    ${TEAM_GSA_EMAIL?} \
--member="serviceAccount:${PROJECT_ID?}.svc.id.goog[cnrm-system/cnrm-controller-manager-${NAMESPACE?}]" \
    --role="roles/iam.workloadIdentityUser" \
    --project ${PROJECT_ID?}

# grant permission for monitoring
gcloud projects add-iam-policy-binding ${PROJECT_ID?} \
    --member="serviceAccount:${TEAM_GSA_EMAIL?}" \
    --role="roles/monitoring.metricWriter"

# grant the ASO controller to use this GCP service account.
WORKLOAD_IDENTITY_POOL="${PROJECT_ID?}.svc.id.goog"  # don’t change this
ASO_NAMESPACE=kontrollers-azureserviceoperator-system  # don’t change this
ASO_KSA=azureserviceoperator-default  # Don’t change
gcloud iam service-accounts add-iam-policy-binding ${TEAM_GSA_EMAIL?} \
 --role roles/iam.workloadIdentityUser \
 --member "serviceAccount:${WORKLOAD_IDENTITY_POOL?}[${ASO_NAMESPACE?}/${ASO_KSA?}]" \
 --condition None

```

2. Create a new kubernetes namespace for the user team, then assign the GCP and Azure identity to the namespace:

```shell
cat <<EOF > /tmp/aks_context.yaml
# Create a namespace for user team
apiVersion: v1
kind: Namespace
metadata:
  name: "${NAMESPACE?}"
---
# Config this namespace for KCC
apiVersion: core.cnrm.cloud.google.com/v1beta1
kind: ConfigConnectorContext
metadata:
  name: configconnectorcontext.core.cnrm.cloud.google.com
  namespace: "${NAMESPACE?}"
spec:
  billingProject: "${PROJECT_ID?}"
  googleServiceAccount: "${TEAM_GSA_EMAIL?}"
  requestProjectPolicy: BILLING_PROJECT
---
# Config this namespace for composition
apiVersion: composition.google.com/v1alpha1
kind: Context
metadata:
  name: context
  namespace: "${NAMESPACE?}"
spec:
  project: "${PROJECT_ID?}"
EOF

kubectl apply  -f /tmp/aks_context.yaml
```

### User Team

#### **Step 1: Create the AKS cluster from the composition**

The following will create an instance of the composition defined by the platform team in step 1\. The composition will create these resources: 

- An Azure resource group  
- An AKS cluster  
- A GCP attached cluster object

```shell
AKS_NAME= <name>  # The name you want to give the attached cluster.
NAMESPACE= <namespace>  # The same namespace as was used by the platform team in step 2A
ATTACHED_REGION= <region>  # The GCP region in which to create the attached cluster
K8S_VERSION= <version>  # The Kubernetes version to run on AKS (e.g. "1.29")
ATTACHED_VERSION= <version>  # The Attached Clusters platform version to use (see below how to get a list of available versions).
AZURE_REGION= <region>  # The same region used in Chapter 3
PROJECT_NUMBER= <project>  # GCP project number
ADMIN_USER= <user@example.com>  # A user that will have cluster admin privileges
cat <<EOF > /tmp/attached_cr.yaml
apiVersion: idp.mycompany.com/v1
kind: AttachedAKS
metadata:
  name: "${AKS_NAME?}"
  namespace: "${NAMESPACE?}"
spec:
  gcpRegion: "${ATTACHED_REGION?}"
  kubernetesVersion: "${K8S_VERSION?}"
  attachedPlatformVersion: "${ATTACHED_VERSION?}"
  azureRegion: "${AZURE_REGION?}"
  gcpProjectNumber: "${PROJECT_NUMBER?}"
  adminUsers:
  - "${ADMIN_USER?}"
EOF

kubectl apply -f /tmp/attached_cr.yaml
```

You can view the list of supported GCP regions for Attached clusters [here](https://cloud.google.com/kubernetes-engine/multi-cloud/docs/attached/eks/reference/supported-regions).

You can get the list of all supported Attached platform versions by running:

```shell
gcloud container attached get-server-config --location=${ATTACHED_REGION?}
```

Users can create multiple `AttachedAKS` CRs with different names to create multiple AKS Attached clusters.

You can verify the AKS cluster has been created by running: `kubectl get managedclusters -n ${NAMESPACE?}`

Note: After this point the AKS cluster should be ready, but it will not be attached yet. Proceed to the next step to complete the attach process.

#### **Step 2: Apply the attached manifest to the AKS cluster**

Although the AttachedCluster GCP resource is created by the composition, the cluster is still not attached to GCP because the attached manifest is not installed in the AKS cluster. This is an agent that dials back to GCP to manage the attached cluster. It can’t be installed by KCC right now because most production kubernetes clusters are private clusters. The KCC cluster has no way to connect to these clusters. So it requires a manual step to install the attached manifests.

GCP provides a `gcloud` command to generate the attached manifest for the attached clusters. Users need to connect to this AKS cluster and apply the manifest.

Here  are the steps.

```shell
NAMESPACE= ... # The namespace used for the AttachedAKS CR in step 1
AKS_NAME=$(kubectl get AttachedAKS -n ${NAMESPACE?} \
  -o=jsonpath='{.items[0].metadata.name}')
ATTACHED_REGION=$(kubectl get AttachedAKS -n ${NAMESPACE?} \
  -o=jsonpath='{.items[0].spec.gcpRegion}')
ATTACHED_VERSION=$(kubectl get AttachedAKS -n ${NAMESPACE?} \
  -o=jsonpath='{.items[0].spec.attachedPlatformVersion}')

# Generate the attached install manifest
gcloud container attached clusters generate-install-manifest \
  $AKS_NAME \
  --location=${ATTACHED_REGION?} \
  --platform-version ${ATTACHED_VERSION?} \
  --output-file=/tmp/install-agent-${AKS_NAME?}.yaml

# Change context to the AKS cluster
az aks get-credentials --name ${AKS_NAME?}-aks \
  --resource-group ${AKS_NAME?}-rg --file /tmp/kubeconfig-aks

# Apply the attached manifest
KUBECONFIG=/tmp/kubeconfig-aks kubectl apply -f /tmp/install-agent-${AKS_NAME?}.yaml
```

#### **Step 3: Verify the cluster is attached**

You can verify a successful cluster attach using the following methods:

3. `gcloud container attached clusters describe ${AKS_NAME?} --location $ATTACHED_REGION`

A detailed block of information about the attached resource will be returned.

4. `kubectl get ContainerAttachedCluster -n ${NAMESPACE?}`  
   The resource will be listed as **Ready** with a status of **UpToDate**.

#### **Step 4: Connect to the cluster**

Once the cluster is attached, you can connect to it by running the following command:

```shell
# Obtain Google credentials to log into the cluster using the current gcloud user:
gcloud container attached clusters get-credentials ${AKS_NAME?} \
  --location=${ATTACHED_REGION?} 

# Get cluster information to verify authentication works
# and the cluster is properly attached:
kubectl cluster-info
```

#### **Step 5: Clean up Azure resources**

Switch your kubectl context back to the Config Controller context, then run:

```shell
kubectl delete attachedakses.idp.mycompany.com ${AKS_NAME?} -n ${NAMESPACE?}

# verify deletion of attached cluster
kubectl get containerattachedcluster -n ${NAMESPACE?}
# verify deletion of resource group
kubectl get ResourceGroup.resources.azure.com -n ${NAMESPACE?}
```
