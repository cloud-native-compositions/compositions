# Scenario: App team project creation

The intent of this scenario is to show how an infrastructure team can use Compositions to create a new KCC namespace, and an associated GCP project for an application team. 

## Facade

The composition uses `AppTeam` CRD as its Facade (input schema). The facade CRD was created using [kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api). A developer using the `AppTeam`  CRD would not need to have any  knowledge of how the underlying KCC objects are defined. 

```py
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  ...
  name: appteams.idp.mycompany.com
spec:
  ...
    kind: AppTeam
  ...
    schema:
      openAPIV3Schema:
  ...
            properties:
              project:  
                type: string
              adminUser:
                type: string
              billingUser:
                type: string
              folder:
                type: string
            required:
            - project
            - adminUser
            - billingAccount
            - folder
 ...
```

## Composition definition

The Composition and its corresponding Facade definition can be found [here](https://github.com/cloud-native-compositions/compositions/blob/main/samples/AppTeam/composition/appteam.yaml).   
A preview of the `appteams` composition definition:

```py
apiVersion: composition.google.com/v1alpha1
kind: Composition
metadata:
  name: appteams
  namespace: default
spec:
  inputAPIGroup: appteams.idp.mycompany.com # inputAPIGroup specifies the CRD from which values are accepted as inputs into  this composition
  namespaceMode: explicit # composition contains a mix of namespace-scoped and cluster-scoped resources
  expanders:              # list of expander-stages
  - type: jinja2          # stages uses Jinja2 expander. When using CEL, expander template use CEL syntax and be limited to CEL features.
    version: v0.0.1
    name: project      # Stage name
    template: |        # Template using Jinja2 syntax and features
      {% set managedProject = appteams.spec.project %}
      apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
      kind: Project
      metadata:
        name: {{ managedProject }}
        namespace: {{ appteams.metadata.namespace }}
        labels:
          createdby: "composition-appteam"
      ...
  - type: jinja2      
    name: namespace   # second stage called `namespace`
    template: |
      ... 
  - type: jinja2
    name: setup-kcc   # third stage called setup-kcc.
    template: |
      ...
   ... # more stages
```

## Using the Composition 

### Apply the composition

Clone the [repo](https://github.com/cloud-native-compositions/compositions) to get the samples. Apply the `composition/appteam.yaml` to create the Composition and its corresponding facade CRD.

```
git clone https://github.com/cloud-native-compositions/compositions.git

cd samples/AppTeam
kubectl apply -f composition/appteam.yaml
```

Check if the composition is installed successfully. You should see something like this.

```
❯ kubectl get composition appteams -o json | jq .status
{
  "stages": {
    "bucket": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "compositions-context": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "namespace": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "project": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "project-owner": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    },
    "setup-kcc": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    }
  }
}
```

### Using the Composition

The composition can be used by creating an instance of the `AppTeam` CRD. 

```py
TEAM_NAME=team-$(tr -dc a-z </dev/urandom | head -c 6)
GCP_FOLDER=  # "00000000000" Set this
GCP_BILLING= # "000000-000000-000000" Set this
ADMINUSER=  # someuser@company.com Set this

# Create a GCP project with name "team-$randomSuffix"
kubectl apply -f - <<EOF
apiVersion: idp.mycompany.com/v1alpha1
kind: AppTeam
metadata:
  name: ${TEAM_NAME?}
  namespace: config-control
spec:
  project: ${TEAM_NAME?}
  # A human who needs access to the project in cloud console
  adminUser: ${ADMINUSER?}
  # Please change this to your billing account to be associated with the project
  billingAccount: ${GCP_BILLING?}
  # Set this to the appropriate folder for the project to be created in
  folder: "${GCP_FOLDER?}"
EOF
```

Composition-engine reconciles the  `AppTeam` CR and performs the  following actions:

1. Creates a new Google Cloud project  
2. Creates a kubernetes namespace for the new project in the Config Controller kubernetes cluster  
3. Enables Config Connector (KCC) in this namespace, configuring it to create resources in the GCP project, using the GCP ServiceAccount  
4. Creates a GCP ServiceAccount that KCC will use when managing resources in the project  
5. Creates an IAMPartialPolicy to allow KCC's Kubernetes Service Account to use the Google Cloud ServiceAccount  
6. Grants KCC ownership of this project  
7. Creates a storage bucket in the project for use by the application team. This step also verifies everything is set up correctly.

Changing the AppTeam facade or the Composition objects, triggers a reconcile and the controller updates the KCC objects.

### Verify

Verify the KCC resources are created and reconciled successfully. Look at the `READY` and `STATUS` columns of the command output. They should be `True` and `UpToDate` respectively.

```
# KCC objects in config-control namespace
kubectl get iamserviceaccount kcc-${TEAM_NAME?} -n config-control
kubectl get iampartialpolicy -n config-control\
 ${TEAM_NAME?}-sa-workload-identity-binding
kubectl get iampartialpolicy -n config-control kcc-owners-permissions-${TEAM_NAME?}
kubectl get project ${TEAM_NAME?} -n config-control

# KCC objects in the new namespace
kubectl get storagebucket -n ${TEAM_NAME?} test-bucket-${TEAM_NAME?}
kubectl get configconnectorcontext -n ${TEAM_NAME?}


# A helper script includes the above commands:
./get_appteam.sh ${TEAM_NAME?}
```

Also inspect the `Plan` object created for the `appteams` instance. This is useful to debug if something goes amiss:

```
# Plan object corresponding to the appteam instance
# The plan object name is obtained by joining 'facade crd' and 'cr.name'
kubectl get plan -n config-control appteams-${TEAM_NAME?} -o yaml | less
```

The Plan’s spec records the expanded objects for each stage. The status details the applied object count as well as the health of the applied KCC objects.  