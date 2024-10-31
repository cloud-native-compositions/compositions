# Overview

## Goals
Help infrastructure platform teams build self-service developer platforms that work consistently across clouds. We are doing this by creating a K8s-native way of orchestrating cloud services and kubernetes resources.


## Customer Context
we have found that our customers often use  ACK/ASO/KCC as an input into a custom-built internal developer platform (IDP). In nearly all cases, platform teams are maintaining a custom orchestration layer between the developer interface and the cloud resources. The orchestration layer is typically responsible for: 
* RBAC
* Resource creation ordering
* Continuous reconciliation
* Hydration of templates with developer inputs
* Security checks
* Automated and manual approval workflows
* Abstracting platform complexity for developer ease-of-use 

There is no consistent K8s-native set of tooling used to build this orchestration layer. The most common solution is to write a layered set of Helm charts, with the top layer exposing only the parameters available to developers. Existing solutions tend to be challenging to maintain and require additional investment in hydration pipelines and custom RBAC models. Customers have been asking us to help address the challenges of orchestrating across different cloud services and teams / roles. The solution must be flexible/extensible so that existing solutions can be “plugged in”, but relatively simple to use out-of-the box in a greenfield scenario. Additionally, our customers have been asking for a consistent approach to cloud resource orchestration across cloud service vendors.

## Product Description
Compositions provides platform administrators with a framework for building  secure, qualified and reusable APIs for application teams to consume cloud resources in a self-service manner. Compositions can be used as the orchestration layer below an IDP, or they can be used directly by developers either through the K8s CLI or through a GitOps tool such as Config Sync or Argo CD. 

### How Compositions Work
Compositions aims to provide a K8s-based solution for the problem of self-service infrastructure management. It is a set of Kubernetes APIs (CRDs) and controllers that allow any K8s resources, including KCC/ASO/ACK resources, to be logically grouped together and consumed as a single resource (a “composition”). By building compositions as K8s-native resources, compositions can use K8s features such as RBAC and continuous reconciliation. 

Compositions will be built as OSS and will be available as a managed service through Config Controller. This private preview will use Config Controller to install all the components necessary to build compositions across Google Cloud, AWS, and Azure. 

Resources available to be managed through compositions by default are any resources available through KCC, ACK, and ASO, as well as all K8s-native constructs (such as namespaces and pods). KCC, ACK, and ASO provide mappings to Google Cloud, AWS, and Azure APIs respectively to Kubernetes - think of these as the building blocks for compositions. Compositions allow you to build your own abstractions that consume these building blocks. Compositions can also manage resources provided by other K8s operators, if you wish to bring your own operators. 

### Design Principles
Some of the design principles we aim to adhere to are:
* For life cycle management of cloud resources, delegate to lower level controllers for individual cloud resources.
* Consistent experience across clouds.
* Extensible architecture for meeting customer choice of configuration language, tooling.
* Well defined core for quick development and Day 1 usage.
* Automate, yet provide control and visibility.
* Compatibility with K8s ecosystem tooling, RBAC, gitops workflows, support shift-left patterns etc.

### Overall Architecture
The below diagram captures the overall architecture of compositions.

![image](architecture.png)

### Compositions Design
A composition is a YAML file that defines a set of resources and how they relate to each other. The key concepts and APIs that will be used in this private preview are described below.

#### Key Concepts:

1. **Composition** \- Defined by a platform admin, a composition is a YAML file that defines a set of resources and any logic (such as ordering, loops, cross-referencing) required to create a service that can be used by app development teams. A composition has:  
   1. A list of Expanders  
      1. Each expander is a stage of the Composition.  
      2. Expanders are run in list order (sequencing)  
      3. Each expander has a template (inline template string)  
      4. (inline) ValuesFrom to specify where dependent values are pulled from  
   2. The input Facade API name  
2. **Facade APIs** \- the Facade API is a CRD that acts as the developer interface for a composition. A platform admin would define the Facade API to expose only the fields in a composition that they want developers to have access to. Developers can use the Facade API for self-service access to the service(s) defined in the composition.   
3. **Expander** \- Expander is responsible for converting composite resources into individual managed resources. It is a service which takes in some YAML and expands it into a different set of YAML. The platform admin will configure a composite resource to be acted on by a particular expander. The expander will take the configuration provided by the platform admin and the input yaml from the developer (through the Facade) and generate the output yaml. The output yaml would be used to instantiate the managed resources.  
   1. Introducing the concept of an expander means compositions can use any templating tool that is able to take YAML as input and give YAML as output. This preview uses Jinja2 as the expander, but other templating tools could be used instead of, or in addition to, Jinja2. For example, Helm, CEL, and Kustomize could be used as expanders. Since all expanders have the same input and output language, they can be easily chained.  
   2. The composition can leverage any features of the expanders it uses. For example, Jinja2 supports loops, so a composition using Jinja2 can also use loops.   
4. **Getters** \- A getter is used to read values from kubernetes resources and use them as parameters when evaluating stages. It allows the composition to reference resources/fields that exist in the cluster but are not defined by the composition.   
5. **Stages** \- Stages can be used within a composition to define the order in which resources should be reconciled. A stage is nothing but an expander in the Composition. Any resources defined in an earlier block will be reconciled before any resources in a later block will attempt to be reconciled. Multiple expanders can be used within the same block (for example, the Cloud SQL scenario in Chapter 4 has a block that uses both Jinja2 and Getter expanders).    
6. **Sequencing** \- in addition to using blocks, resource ordering can be done by defining rules for how resource creation should be ordered. In the Cloud SQL scenario (chapter 4), we show how a resource can be made to wait on the availability of a service account email before being reconciled. 
#### APIs (CRDs):
Compositions consist of the following user-facing APIs:

* **Composition**: This is the API used by the platform admin to define a composition. It contains a set of Expander sections with associated config/templates. These are evaluated in response to Input/Facade APIs to generate KRM resources.
* **Input/Facade API**: CRD that is defined by the platform admin. Provides values for associated composition. It is used by developers for self-service access to the services defined in a composition. 

## Assumed Knowledge
This private preview assumes knowledge of Kubernetes (K8s), K8s CRDs, and KCC, ACK, ASO is assumed.

