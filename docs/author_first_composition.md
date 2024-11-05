# Writing your first composition

This document shows you how to write a simple composition that manages K8s-native resources such as a Deployment and ConfigMap.

## Concepts

The concepts explored in this chapter are:

* Composition  
  * Stages  
  * Expander types  
* Facade

## Your First Composition

We will start with a simple Composition that deploys a web-server and a landing page that lists the members of a team.

Typically to achieve this a team would need to deploy:

1. Web-server Pod  
2. Service connecting to the pod  
3. Configmap that creates a static web page in the web-server pod

Using Compositions, the same team would need to create a single object that looks something like this:

```yaml
apiVersion: idp.mycompany.com/v1alpha1
kind: TeamPage
metadata:
  name: teampage
  namespace: team-sales
spec:
  members:
  - name: Jo
    role: Eng Manager
  - name: Jane
    role: Lead
  - name: Bob
    role: Developer
```

Such an object that abstracts away the underlying complexity and provides the input to a Composition is called a Facade. A Facade is a CRD that captures the inputs from the user which are then used in the Composition.

## Composition

Let’s define a Composition that uses the Facade and creates a web-server Deployment, Service and a Configmap. We shall use `jinja2` based templating in this example.   
   
A preview of such a [composition](https://github.com/cloud-native-compositions/compositions/tree/main/samples/FirstComposition) looks like this:

```yaml
apiVersion: composition.google.com/v1alpha1
kind: Composition
metadata:
 name: team-page
spec:
 # we have a plan to replace  inputAPIGroup field with apiVersion, Kind fields.
 # inputAPI:
 #   apiVersion: idp.mycompany.com/v1alpha1
 #   kind: TeamPage
 inputAPIGroup: teampages.idp.mycompany.com    # Facade API
 expanders:
 - type: jinja2  # pluggable expander
   name: server  # stage
   template: |   # jinja2 template
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        # Use the teampages Facade's name
        name: team-{{ teampages.metadata.name }}
        # The namespace is set to the facade's namespace
        namespace: default
        labels:
          # use facade's name in the label
          app: nginx-{{ teampages.metadata.name }}
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: nginx-{{ teampages.metadata.name }}
        template:
          metadata:
            labels:
              # use facade name in the pod's label
              app: nginx-{{ teampages.metadata.name }}
          spec:
            containers:
              - name: server
                image: nginx:1.16.0
                ports:
                  - name: http
                    containerPort: 80
                    protocol: TCP
                volumeMounts:
                  - name: index
                    mountPath: /usr/share/nginx/html/
            volumes:
              - name: index
                configMap:
                  # use the configmap created by this composition
                  name: team-{{ teampages.metadata.name }}-page
      ---
      apiVersion: v1
      kind: Service
      metadata:
        # include the facade name in the service name
        name: team-{{ teampages.metadata.name }}-landing
        namespace: default
        labels:
          app: nginx-{{ teampages.metadata.name }}
      spec:
        ports:
        - port: 80
          protocol: TCP
        selector:
          # match the web-server pod
          app: nginx-{{ teampages.metadata.name }}
      ---
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: team-{{ teampages.metadata.name }}-page
        namespace: default
      data:
        index.html: |
           <html>
           <h1>{{ teampages.metadata.name  }}</h1>
           <table>
             <tr>
               <th>Name</th>
               <th>Role</th>
             </tr>
           {% for member in teampages.spec.apps %}
             <tr>
               <td>{{ member.name }}</td>
               <td>{{ member.role }}</td>
             </tr>
           {% endfor %}
           </table>
           </html>
```

The `team-page` Composition

* Uses `teampages.idp.mycompany.com` CRD as the input schema  
* Has single stage with 3 resources  
* Uses expander of type `jinja2`

Please note that the templating language is [pluggable](https://github.com/cloud-native-compositions/compositions/tree/main/expanders/helm-expander). The Composition is split into expander stages. Each expander is evaluated and applied before the next stage is processed. This introduces implicit sequencing.  In this example, we will create all the objects in a single expander stage.

##  Creating Input Schema using kubebuilder

The composition uses `TeamPage` CRD as its Facade (input schema). The facade CRD can be  created using [kubebuilder](https://book.kubebuilder.io/cronjob-tutorial/new-api).  Refer [Composition authoring (Step 3: Write the Facade API)](authoring_walkthrough.md) for steps to use kubebuilder to create a Facade CRD.

Create scaffolding for the `TeamPage` CRD

```shell
kubebuilder create api --group idp --version v1alpha1 --kind TeamPage --controller=false --resource
```

Edit the golang file to add the `spec` fields

```shell
# choose editor of your choice
vim api/v1alpha1/teampage_types.go
```

Edit the `Spec` struct to include the team members:

```c
## Edit the Spec struct to include the team members
type Member struct {
    Name string `json:"name"`
    Role string `json:"role"`
}

## The TeamPageSpec should look like this
type TeamPageSpec struct {
    Members []Member `json:"members"`
}
```

The Generated CRD should look like this:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  ...
  name: teampages.idp.mycompany.com
spec:
  ...
    kind: TeamPage
  ...
    schema:
      ...
```

## Apply the composition

Clone the [repo](https://github.com/cloud-native-compositions/compositions) to get the samples. Apply the `composition/teampage.yaml` to create the Composition and its corresponding facade CRD.

```shell
# if not cloned already
git clone https://github.com/cloud-native-compositions/compositions.git

cd samples/FirstComposition
kubectl apply -f composition/teampage.yaml
```

Check if the composition is installed successfully. You should see something like this.

```shell
❯ kubectl get composition team-page -o json | jq ".status"
{
  "stages": {
    "server": {
      "reason": "ValidationPassed",
      "validationStatus": "success"
    }
  }
}
```

## Using the composition

Once the Composition and the Facade which specifies the input schema are installed, the composition is ready for use. A team in your org creates an instance of the `TeamPage` Facade as an input to the composition.

The first step is to create a namespace for the team. This namespace is where the `TeamPage`  would be created.

```shell
export NAMESPACE=my-team
kubectl create namespace ${NAMESPACE?}
```

Next we create a `TeamPage` instance in the namespace:

```shell
kubectl apply -f - <<EOF
apiVersion: idp.mycompany.com/v1alpha1
kind: TeamPage
metadata:
  name: landing
  namespace: ${NAMESPACE?}
spec:
  members:
  - name: Jo
    role: Eng Manager
  - name: Jane
    role: Lead
  - name: Bob
    role: Developer
EOF
```

## Verify

Verify the k8s objects are created successfully:

```shell
kubectl get deployment -n ${NAMESPACE?}
kubectl get service -n ${NAMESPACE?}
kubectl get configmap -n ${NAMESPACE?}
```

Check the web server:

```shell
kubectl port-forward service/team-landing-landing -n my-team 5555:80

# seperate terminal, you should see the web page being served correctly
❯ curl localhost:5555
<html>
<h1>landing</h1>
<table>
  <tr>
    <th>Name</th>
    <th>Role</th>
  </tr>

  <tr>
    <td>Jo</td>
    <td>Eng Manager</td>
  </tr>

  <tr>
    <td>Jane</td>
    <td>Lead</td>
  </tr>

  <tr>
    <td>Bob</td>
    <td>Developer</td>
  </tr>

</table>
</html>
```
