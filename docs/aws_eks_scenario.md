# Scenario: EKS cluster creation & attach

This scenario will demonstrate the multi-cloud capabilities of compositions by leveraging ACK to manage AWS resources. We use the example of creating an AWS EKS cluster and attaching it to GKE Enterprise.

## Deploy an attached EKS cluster with a composition

With KCC compositions, users can separate tasks across platform teams and user teams. The platform team defines the composition with all the dependencies of the attached EKS cluster. Subsequently, the user team can use the composition to provision an attached EKS cluster based on their needs. Below is an example of this flow for an attached EKS cluster. (The complete sample is uploaded in [github](https://github.com/cloud-native-compositions/compositions/samples/AttachedEKS). We will go through it in this guide.)

### Platform Team

#### **Step 1: Create the composition for the attached EKS cluster**

First, the platform administrator creates the composition as follows:

```shell
kubectl apply -f \
https://raw.githubusercontent.com/cloud-native-compositions/compositions/main/samples/AttachedEKS/AttachedEKS/01-composition.yaml
```

This file contains a CRD, which is an interface for the composition. The user team will input the variables defined in this CRD for the EKS cluster.

The file also contains the composition itself. The composition is a combination of the following resources:

- An AWS VPC  
- An AWS internet gateway  
- An AWS public route table and a private route table  
- Multiple AWS public and private subnets  
- An AWS elastic IP  
- An AWS NAT gateway  
- An AWS cluster role and a nodegroup role  
- An EKS cluster  
- An AWS EKS node group  
- An AWS EKS access entry   
- A Kubernetes configMap to store the EKS issuer URL  
- An AWS field export to export the EKS issuer URL to the configMap  
- A GCP Attached Cluster resource

#### **Step 2: Create a separate identity for the user team** 

The platform admin can create separate identities for different user teams to isolate the identity and permissions for different teams. A summary of the steps are:

1. Create a new GCP service account for the user team.  
2. Create a new kubernetes namespace for the user team.   
3. Assign the GCP identity to the namespace.

After these steps, when the user team creates resources in their kubernetes namespace, the resources will be created with the new designated identity.

Here are the details for how to create the separate identity:

1. Create a new GCP service account for the user team:

```shell
export TEAM_NAME=<a name for the user team> # example: team-2
export NAMESPACE=${TEAM_NAME?} # you can choose a different name
export TEAM_GCP_SA_NAME="${TEAM_NAME?}" # you can choose a different name
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
```

2. Create a new kubernetes namespace for the user team, then assign the GCP identity to the namespace:

```shell
cat <<EOF > /tmp/eks_context.yaml                                                                                                                                         
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

kubectl apply  -f /tmp/eks_context.yaml
```

### User Team

#### **Step 1: Create the EKS cluster using the composition**

The following will create an instance of the composition defined by the platform team in step 1 which will create these resources: 

- An AWS VPC  
- An AWS internet gateway  
- An AWS public route table and a private route table  
- Two AWS public and private subnets  
- An AWS elastic IP  
- An AWS NAT gateway  
- An AWS cluster role and a nodegroup role  
- An EKS cluster  
- An AWS EKS node group  
- An AWS EKS access entry   
- A kubernetes configMap to store the EKS issuer URL  
- An AWS field export to export the EKS issuer URL to the configMap  
- A GCP attached cluster object

```shell
EKS_NAME= <name>  # The name you want to give the attached cluster.
NAMESPACE= <namespace>  # The same namespace as was used by the platform team.
ATTACHED_REGION= <region>  # The GCP region to create the attached cluster in (see below for supported regions).
K8S_VERSION= <version>  # The Kubernetes version to on the EKS cluster (e.g. 1.29).
ATTACHED_VERSION= <version>  # The Attached Clusters platform version to use (see below how to get a list of available versions).
AWS_REGION= <region>  # The region where the EKS cluster will run.
PROJECT_NUMBER= <project>  # The GCP project number where the cluster will be attached.
ADMIN_USER= <user@example.com  # A Google user that will have cluster admin privileges. Typically, the user you are logged in into the gcloud tool as. 
AWS_USER_ARN= <arn> # The ARN of an AWS user to register as cluster owner (looks like "arn:aws:iam::000000000000:user/username").
cat <<EOF > /tmp/attached_eks_cr.yaml
apiVersion: idp.mycompany.com/v1
kind: AttachedEKS
metadata:
  name: "${EKS_NAME?}"
  namespace: "${NAMESPACE?}"
spec:
  gcpRegion: "${ATTACHED_REGION?}"
  kubernetesVersion: "${K8S_VERSION?}"
  attachedPlatformVersion: "${ATTACHED_VERSION?}"
  awsRegion: "${AWS_REGION?}"
  gcpProjectNumber: "${PROJECT_NUMBER?}"
  adminUsers:
  - "${ADMIN_USER?}"
  awsAccessIdentity: "${AWS_USER_ARN?}"
  awsAvailabilityZones:
  - zoneNameSuffix: b
    publicSubnet: "10.0.11.0/24"
    privateSubnet: "10.0.1.0/24"
  - zoneNameSuffix: c
    publicSubnet: "10.0.12.0/24"
    privateSubnet: "10.0.2.0/24"
EOF

kubectl apply -f /tmp/attached_eks_cr.yaml
```

You can view the list of supported GCP regions for Attached clusters [here](https://cloud.google.com/kubernetes-engine/multi-cloud/docs/attached/eks/reference/supported-regions).

You can get the list of all supported Attached platform versions by running:

```shell
gcloud container attached get-server-config --location=${ATTACHED_REGION?}
```

Users can create multiple `AttachedEKS` CRs with different names to create multiple EKS Attached clusters.

Note: After this point the EKS cluster should be created, but it will not be attached yet. Proceed to the next step to complete the attach process.

#### **Step 2: Apply the attached manifest to the EKS cluster**

Although the AttachedCluster GCP resource is created by the composition, the cluster is still not attached to GCP because the attached *install manifest* is not installed in the EKS cluster. This is an agent that dials back to GCP to manage the attached cluster. It can’t be installed by KCC right now because most production kubernetes clusters are private clusters. The KCC cluster has no way to connect to these clusters. So it requires a manual step to install the attached manifests.

GCP provides a `gcloud` command to generate the attached manifest for the attached clusters. Users need to connect to this EKS cluster and apply the manifest.

Here  are the steps:

```shell
NAMESPACE= ... # The namespace used for the AttachedEKS CR in step 1
EKS_NAME=$(kubectl get AttachedEKS -n ${NAMESPACE?} -o=jsonpath='{.items[0].metadata.name}')
ATTACHED_REGION=$(kubectl get AttachedEKS -n ${NAMESPACE?} -o=jsonpath='{.items[0].spec.gcpRegion}')
ATTACHED_PLATFORM_VERSION=$(kubectl get AttachedEKS -n ${NAMESPACE?} -o=jsonpath='{.items[0].spec.attachedPlatformVersion}')
AWS_REGION=$(kubectl get AttachedEKS -n ${NAMESPACE?} -o=jsonpath='{.items[0].spec.awsRegion}')

# Generate the attached manifest
gcloud container attached clusters generate-install-manifest \
  ${EKS_NAME?} \
  --location=${ATTACHED_REGION?} \
  --platform-version ${ATTACHED_PLATFORM_VERSION?} \
  --output-file=/tmp/install-agent-${EKS_NAME?}.yaml

# Change context to the EKS cluster (if it says the cluster status is CREATING,
# wait a bit and try again)
aws eks update-kubeconfig --name ${EKS_NAME?}-cluster --region ${AWS_REGION?} \
  --kubeconfig /tmp/kubeconfig-eks

# Apply the attached manifest
KUBECONFIG=/tmp/kubeconfig-eks kubectl apply -f /tmp/install-agent-${EKS_NAME?}.yaml
```

#### **Step 3: Verify the cluster is attached**

You can verify a successful cluster attach using the following methods:

1. `gcloud container attached clusters describe ${EKS_NAME?} --location ${ATTACHED_REGION?}`

A detailed block of information about the attached resource will be returned.

2. `kubectl get ContainerAttachedCluster -n ${NAMESPACE?}`  
   The resource will be listed as **Ready** with a status of **UpToDate**.

#### **Step 4: Connect to the cluster**

Once the cluster is attached, you can connect to it by running the following command:

```shell

# Obtain Google credentials to log into the cluster using the current gcloud user:
gcloud container attached clusters get-credentials ${EKS_NAME?} \
  --location=${ATTACHED_REGION?}

# Get cluster information to verify authentication works and cluster is properly attached:
kubectl cluster-info
```

#### **Step 5: Clean up AWS resources**

Switch your kubectl context back to the Config Controller context, then run:

```shell
# delete the AWS cluster
kubectl delete attachedekses.idp.mycompany.com ${EKS_NAME?} -n ${NAMESPACE?}
# verify deletion of attached cluster
kubectl get containerattachedcluster -n ${NAMESPACE?}
# verify deletion of the EKS cluster (this may take a few minutes)
kubectl get cluster.eks.services.k8s.aws -n ${NAMESPACE?}
# verify deletion of the VPC (this may take a few minutes)
kubectl get vpc.ec2.services.k8s.aws -n ${NAMESPACE?}
```
