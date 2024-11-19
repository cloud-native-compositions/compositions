# Scenario: Cloud SQL creation

The intent of this scenario is to show how an infrastructure team can create a composition that the application teams can use to create CloudSQL in a high-availability(HA) mode. The application team creates a single `CloudSQL` custom resource (facade instance). The backing composition creates SQL instances, IAM service accounts, IAM roles, KMS Keyrings, KMS keys and sets it up with a single master and multiple replicas.  

In later sections, we will use this composition to show how more advanced compositions can be written. For now, we are focusing on how to use an off-the-shelf composition.

The platform admin will create a composition called `sqlha`, which defines all resources required to create a highly-available Cloud SQL instance. They will also define a Facade API that exposes only the CloudSQL fields they want a developer to specify. A developer will then use this Facade API to create a working Cloud SQL instance without needing to understand the individual kubernetes resources required, nor even to understand the underlying GCP APIs (such as IAM service accounts, IAM roles, and Cloud SQL APIs). 

## Facade

The composition uses `CloudSQL` CRD as its Facade (the developer interface / input). The facade CRD was created using [kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api) (this will be explained in detailed in Chapter 8). A developer using the `CloudSQL`  CRD would not need to have any knowledge of how the underlying KCC objects are defined. 

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  ...
  name: cloudsqls.idp.mycompany.com
spec:
  ...
    kind: CloudSQL
  ...
    schema:
      ...
```

## Composition definition

The composition that corresponds to the Facade  
A preview of the `sqlha` composition definition:

```yaml
apiVersion: composition.google.com/v1alpha1
kind: Composition
metadata:
  name: sqlha
  namespace: default
spec:
  inputAPIGroup: cloudsqls.idp.mycompany.com # inputAPIGroup specifies the CRD from which values are accepted as inputs into  this composition
  expanders:              # list of expander-stages
  - type: jinja2          # stages uses Jinja2 expander. When using CEL, expander template use CEL syntax and be limited to CEL features.
    version: v0.0.1
    name: enable-services      # Stage name
    template: |        # Template using Jinja2 syntax and features
        ....
  - type: jinja2
    name: create-serviceidentity
    version: v0.0.1
    template: |
      apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
      kind: ServiceIdentity
      ...
  - type: getter # Getter stage to wait for service identity email
    version: v0.0.1
    name: get-serviceidentity
    template: ""
    configref:
      name: sql-siemail
      namespace: default
  - type: jinja2 
    name: sql-instances
    version: v0.0.1
    template: |
      {% for region in cloudsqls.spec.regions %}
      ... # uses the getter values
```

## Using the composition 

### Platform Admin applies the composition

Clone the [repo](https://github.com/cloud-native-compositions/compositions) to get the samples. Apply the `composition/hasql.yaml` to create the Composition and its corresponding facade CRD.

```shell
git clone https://github.com/cloud-native-compositions/compositions.git

cd samples/CloudSQL
kubectl apply -f composition/hasql.yaml
```

Check if the composition is installed successfully. You should see something like this.

```shell
❯ kubectl get composition sqlha -o json | jq .status
{
  "stages": {
    "block2": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "block3": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "enable-services": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "get-serviceidentity": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    }
  }
}
```

### Developer uses the composition

The composition can be used by creating an instance of the `CloudSQL` CRD. 

```shell
NAMESPACE=config-control

kubectl apply -f - <<EOF
apiVersion: idp.mycompany.com/v1alpha1
kind: CloudSQL
metadata:
  name: myteam
  namespace: ${NAMESPACE?}
spec:
  regions:
  - us-east1
  - us-central1
  name: myteam-db
EOF
```

Composition-engine reconciles the  `CloudSQL` CR and performs the  following actions:

1. Enables the required Google Cloud services \- Cloud KMS, IAM, Service Usage, and SQL Admin.   
2. Creates a Service Identity.  
3. Uses the Getter to retrieve the Cloud SQL admin service identity email when it has been created.   
4. Creates a KMSKeyRing  
5. Creates a KMSCryptoKey  
6. Waits for the Getter to retrieve the Cloud SQL admin service identity email before creating an IAMPolicyMember to give the SQL service account access to the previously created KSMCryptoKey.   
7. Creates a SQLInstance as master in the first region.  
8. Creates a SQLInstance as a replica in the second region. Additional regions if present would be set up as replicas.

### Verify

Give it a couple of minutes and then verify if the KCC resources are created and reconciled successfully. Look at the `READY` and `STATUS` columns of the command output. They should be `True` and `UpToDate` respectively.

```shell
# Verify KCC objects exist
kubectl get  serviceidentity -n ${NAMESPACE?}
kubectl get  sqlinstances.sql.cnrm.cloud.google.com -n ${NAMESPACE?}
kubectl get  kmskeyring -n ${NAMESPACE?}
kubectl get  kmscryptokey -n ${NAMESPACE?}
kubectl get  iampolicymember -n ${NAMESPACE?}
kubectl get services.serviceusage.cnrm.cloud.google.com -n ${NAMESPACE?}

# A helper script includes the above commands:
./get_cloudsql.sh ${NAMESPACE?}
```

Sample output:

```shell
❯ ./get_cloudsql.sh ${NAMESPACE?}

ServiceIdentity ----------------------------------------
NAME                      AGE   READY   STATUS     STATUS AGE
sqladmin.googleapis.com   19m   True    UpToDate   19m

SqlInstance --------------------------------------------
NAME                            AGE   READY   STATUS     STATUS AGE
myteam-db-main                  19m   True    UpToDate   15m
myteam-db-replica-us-central1   19m   True    UpToDate   6m57s

KMSKeyRings --------------------------------------------
NAME                           AGE   READY   STATUS     STATUS AGE
kmscryptokeyring-us-central1   19m   True    UpToDate   19m
kmscryptokeyring-us-east1      19m   True    UpToDate   19m

KMSCryptoKeys ------------------------------------------
NAME                           AGE   READY   STATUS     STATUS AGE
kmscryptokey-enc-us-central1   19m   True    UpToDate   19m
kmscryptokey-enc-us-east1      19m   True    UpToDate   19m

IAMPolicyMember ----------------------------------------
NAME                                AGE   READY   STATUS     STATUS AGE
sql-kms-us-central1-policybinding   19m   True    UpToDate   19m
sql-kms-us-east1-policybinding      19m   True    UpToDate   19m

ServiceUsage -------------------------------------------
NAME                          AGE   READY   STATUS     STATUS AGE
cloudkms.googleapis.com       19m   True    UpToDate   19m
iam.googleapis.com            19m   True    UpToDate   19m
serviceusage.googleapis.com   19m   True    UpToDate   19m
sqladmin.googleapis.com       19m   True    UpToDate   19m

```

Also inspect the `Plan` object created for the `cloudsqls` instance. The `Plan` object is an intermediate API that is used to track expanded resources and their status. This is useful to debug if something goes amiss:

```shell
# Plan object corresponding to the appteam instance
#                              >>>facadecrd-cr.name<<<
kubectl get plan -n config-control cloudsqls-myteam -o yaml

apiVersion: composition.google.com/v1alpha1
kind: Plan
...
  name: cloudsqls-myteam
  namespace: config-control
spec:
  stages:
    ...
    enable-services:
      manifest: |2+
        ---
        apiVersion: serviceusage.cnrm.cloud.google.com/v1beta1
        kind: Service
        metadata:
          annotations:
            cnrm.cloud.google.com/deletion-policy: "abandon"
            cnrm.cloud.google.com/disable-dependent-services: "false"
          name: cloudkms.googleapis.com
          namespace: config-control
        spec:
          resourceID: cloudkms.googleapis.com
        ---
        ...
    get-serviceidentity:
      values: '{"identity":{"email":"service-551315786471@gcp-sa-cloud-sql.iam.gserviceaccount.com"}}'
status:
  compositionGeneration: 1
  compositionUID: 88c55c16-590b-4324-aae5-cb3cd19835fb
  conditions:
  - lastTransitionTime: "2024-07-18T11:54:53Z"
    message: 'Evaluated and Applied stages: enable-services, block2, block3'
    reason: ProcessedAllStages
    status: "True"
    type: Ready
  generation: 5
  inputGeneration: 1
  stages:
    block3:
      appliedCount: 8
      lastApplied:
      - group: kms.cnrm.cloud.google.com
        health: Healthy
        kind: KMSKeyRing
        name: kmscryptokeyring-us-east1
        namespace: config-control
        status: Resource is current
        version: v1beta1
      - group: kms.cnrm.cloud.google.com
        health: Healthy
        kind: KMSCryptoKey
        name: kmscryptokey-enc-us-east1
        namespace: config-control
        status: Resource is current
        version: v1beta1
       ...
      resourceCount: 8
    enable-services:
      appliedCount: 4
      lastApplied:
      - group: serviceusage.cnrm.cloud.google.com
        health: Healthy
        kind: Service
        name: cloudkms.googleapis.com
        namespace: config-control
        status: Resource is current
        version: v1beta1
        ...
      resourceCount: 4
```

The Plan’s spec records the expanded objects for each stage. The status details the applied object count as well as the health of the applied KCC objects.