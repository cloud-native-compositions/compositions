# Writing your first GCP composition with KCC resources

This document shows you how to write a simple composition that manages Google Cloud services via KCC.

## Concepts

The concepts explored in this chapter are:

* Context  
* Composition  
  * Stages  
  * Expander types  
* Facade  
* Using compositions to manage GCP resources through KCC

## Your First GCP Composition

Let's look at a Composition that deploys a GCS Bucket. We want to allow easy configuration of the [CORS](https://cloud.google.com/storage/docs/using-cors) policy on the bucket with customizable retention.

An example KCC manifest of such a bucket:

```yaml
apiVersion: storage.cnrm.cloud.google.com/v1beta1
kind: StorageBucket
metadata:
  annotations:
    cnrm.cloud.google.com/force-destroy: "false"
    # StorageBucket names must be globally unique
    name: mybucket-name
    namespace: config-connector
spec:
  lifecycleRule:
  - action:
    type: Delete
    condition:
      age: 10d # delete in 10 days
      withState: ANY
  versioning:
    enabled: true
  uniformBucketLevelAccess: true
  cors:
    - origin: ["http://domain.com"]
      responseHeader: ["Content-Type"]
      method: ["GET", "HEAD", "DELETE"]
      maxAgeSeconds: 3600
```

## Composition

A composition that parameterizes the KCC resource would look like this:

```yaml
apiVersion: composition.google.com/v1alpha1
kind: Composition
metadata:
  name: cors-bucket
spec:
  # we have a plan to replace  inputAPIGroup field with apiVersion, Kind fields.
  # inputAPI:
  #   apiVersion: idp.mycompany.com/v1alpha1
  #   kind: CRBucket
  inputAPIGroup: crbuckets.idp.mycompany.com    # Facade API
  expanders:
  - type: jinja2  # inbuilt jinja2 expander
    name: bucket 
    template: |
       apiVersion: storage.cnrm.cloud.google.com/v1beta1
       kind: StorageBucket
       metadata:
         annotations:
           cnrm.cloud.google.com/force-destroy: "false"
         # StorageBucket names must be globally unique
         # Get the name from the user and prepend the project name
         # user provided name is the facade name
         name: {{ context.spec.project }}-{{ crbuckets.metadata.name }}
         namespace: config-connector
       spec:
         lifecycleRule:
           - action:
               type: Delete
             condition:
               # we want age to be configurable by the user.
               # this will be exposed as a spec field in the facade
               age: {{ crbuckets.spec.retentionDays }}
               withState: ANY
         versioning:
           enabled: true
         uniformBucketLevelAccess: true
         {% if crbuckets.spec.corsURL != '' %}
         cors:
           # URL is exposed as a spec field in the facade
           - origin: ["{{ crbuckets.spec.corsURL }}"]
             responseHeader: ["Content-Type"]
             method: ["GET", "HEAD", "DELETE"]
             maxAgeSeconds: 3600
         {% endif %}
```

The `cors-bucket` Composition

* Uses the following objects:   
  * Facade: Input provided by the app team  
    * `crbuckets.idp.mycompany.com` CRD as the Facade  
  * Context:  `Context` is a built-in object that can be created by the administrator for each namespace. The Context object lets the expanders understand which project they are operating within, so you don’t need to specify it for every resource definition.   
    * `context.spec.project` in the name of the bucket.   
* Has a single stage with 1 resource  
* Uses expander of type `jinja2`

## Apply the composition

Clone the [repo](https://github.com/cloud-native-compositions/compositions) to get the samples. Apply the `composition/cors-bucket.yaml` to create the Composition and its corresponding facade CRD.

```shell
# if not cloned already
git clone https://github.com/cloud-native-compositions/compositions.git

cd samples/FirstGCPComposition
kubectl apply -f composition/cors-bucket.yaml
```

Check if the composition is installed successfully. You should see something like this.

```shell
❯ kubectl get composition cors-bucket -o json | jq ".status"
{
  "stages": {
    "bucket": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    }
  }
}
```

## Defining the Facade

The composition uses `CRBucket` CRD as its Facade. The Facade CRD can be created using [kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api).  Refer [Composition authoring (Step 3: Write the Facade API)](authoring_walkthrough.md) for steps to use kubebuilder to create a Facade CRD.

Create scaffolding for the `CRBucket` CRD

```shell
kubebuilder create api --group idp --version v1alpha1 --kind CRBucket --controller=false --resource
```

Edit the golang file to add the `spec` fields

```shell
# choose editor of your choice
vim api/v1alpha1/crbucket_types.go

## Edit the Spec struct to include cors-url and retention
## The CRBucketSpec should look like this
# type CRBucketSpec struct {
#	CorsURL string `json:"corsURL"`
#	RetentionDays    int   `json:"retentionDays"`
#}
```

The Generated CRD should look like this:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  ...
  name: crbuckets.idp.mycompany.com
spec:
  ...
    kind: CRBucket
  ...
    schema:
      ...
```

## Using the composition

Once the Composition and the Facade which specifies the input schema are installed, the composition is ready for use. A Team in your org creates an instance  of the `CRBucket` to use the composition.

```shell
kubectl apply -f - <<EOF
apiVersion: idp.mycompany.com/v1
kind: CRBucket
metadata:
  name: example-bucket
  namespace: config-control
spec:
  corsURL: "something.foobar.com"
  retentionDays: 10
EOF
```

## Verify

Verify the KCC bucket is created and reconciled successfully:

```shell
kubectl get storagebucket -n config-control
```

Expected output:

```shell
❯ kubectl get storagebucket -n config-control

NAME                                  AGE   READY   STATUS     STATUS AGE
compositions-barni-2-example-bucket   22s   True    UpToDate   21s
```

Also inspect the `Plan` object created for the `crbuckets` instance. This is useful to debug if something goes amiss:

```shell
# Plan object corresponding to the crbuckets instance
#                              >>>facadecrd-cr.name<<<
kubectl get plan -n config-control crbuckets-example-bucket -o yaml | less
```

The Plan’s spec records the expanded objects for each stage. The status details the applied object count as well as the health of the applied KCC object.