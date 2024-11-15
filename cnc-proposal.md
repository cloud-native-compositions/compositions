# Cloud Native Compositions (CNC) Proposal

## Audience
Platform administrators building internal developer platforms (IDPs) based on Kubernetes. The customers of platform administrators are their internal “platform users”.  

## Problem Statement
Many Kubernetes users cite the need to define repeatable patterns for consuming resources, with the ability to monitor and control resource access, usage, and compliance at scale. As more applications are run on Kubernetes, many organizations are now choosing to use K8s as their management platform for all external dependencies (including cloud resources using tools such as ACK, KCC, or ASO), and are building internal developer platforms (IDPs) on top of K8s. This has the benefit of enabling a platform team to provide self-service infrastructure to platform users, while leveraging the K8s ecosystem. 

However, this approach also introduces challenges. Platform teams need to define and implement the contract for the services that are exposed to end users (across many vendor APIs), monitor and control service access and consumption, and attempt to simplify the platform user experience for non-experts in K8s or cloud provider APIs. While there are templating tools that help get to this outcome today (e.g. Helm, Kustomize), there is no **server-side K8s-native solution** that provides platform teams with a consistent way of defining repeatable patterns for how resources such as cloud services should be consumed in their organization. This means that every organization building an IDP today builds it differently, invariably making it costly for them to support it long term. 

Cloud Native Compositions solves for the challenge of how platform teams define services for their K8s-based IDP. It is a framework to define repeatable patterns, called compositions, for how cloud services should be consumed in their organization. Any Kubernetes user can benefit from these K8s-native patterns to reduce context switching even if they are not building a platform, although we are optimizing for the platform use-case.

## Goal
Empower platform teams to build K8s-native internal developer platforms that: 

- Abstract and resolve complex interdependencies between resources, enabling end user self-service for platform teams and simplified resource management for small dev teams
- Provide visibility into how an application is interacting with cloud services
- Are easier and cheaper to build and maintain than existing solutions
- Are compatible with any K8s operator, and existing packaging and templating tools
- Leverage existing extensibility types (CRDs) as the API for interacting with cloud resources

## Non-Goals
We do not want to:
- Build a new templating tool. The framework should be pluggable, such that users can BYO existing templating tools such as CEL, Helm, Jinja etc. 
- Build all of the tooling needed by platform users. CNC will help platform administrators provide services to platform users. The customers of CNC are platform administrators. 
- Define new types of extensibility for Kubernetes besides CRDs.


## Proposal
**Deliverable:** Cloud Native Compositions (CNC) extends Kubernetes to provide platform administrators with a framework for building secure and reusable APIs for application teams to consume in a self-service manner.

**Long-term vision:** The long-term vision of CNC is to help platform teams provide the best experience for developers running applications on Kubernetes. Functionally, CNC will do this by developing a framework as OSS (with a goal to donate to CNCF) for platform teams building internal developer platforms on Kubernetes.

## User Stories
As a **platform administrator**, I want to:
- Define repeatable patterns for consuming cloud resources that can leverage any K8s operator, so that I can more easily define and provision cloud resources at scale in a way that is compatible with the policies and end user needs of my organization.
- Have a consistent way of building (and managing/versioning) custom APIs for end users, so that  end users (e.g. developers, data scientists) in my organization have self-service access to the cloud resources they need. 
- Be able to monitor usage of these custom APIs and the underlying resources they manage.


As a **platform user / downstream user** (e.g. developer, data scientist), I want to:
Easily consume the cloud services I need (without needing to understand how K8s works). 

## Example Use Cases
### Database Provisioning
As a platform administrator, I want to give end users in my organization self service access to SQL databases of different sizes across different cloud providers. I want to define all required resources and logic as a reusable composition, and define an end user interface that exposes only the options I want to be configurable by end users. 

The composition would define:
- Database t-shirt sizes and cloud provider options, and all logic to map these t-shirt sizes to cloud provider APIs.
- A connection config object which can be used by an application and supports region migration

An end user would provision a new SQL database by creating an instance of this composition that defines:
- SQL instance name
- T-shirt size
- Cloud provider
- Cloud region

The end user does not need to be concerned with how their inputs are translated to cloud provider APIs, and does not need to worry about “breaking things” or breaching org policies since their interface and the underlying logic is owned by the platform administrator. 

### K8s Cluster Provisioning
(Using GKE as the example, but equally applicable to any distribution). As a platform administrator, I want to give end users in my organization self service access to K8s clusters. As in the previous example, I want to define all required resources and logic as a reusable composition, and define an end user interface that exposes only the options I want to be configurable by end users. In addition to creating a cluster, we want the composition to deploy administrative workloads and config such as policies, agents etc. 

The composition would define the following resources in Google Cloud (using KCC to provide the mappings from K8s CRDs to Google Cloud APIs):
- GKE cluster
- Container Node Pools
- IAM ServiceAccount
- IAM PolicyMember
- Services (such as cnrm)

The platform administrator would define the end user interface so that an end user can create a new cluster by creating an instance of this composition that defines:
- Cluster name
- Nodepool name
- Max nodes
- Location (e.g. us-east1)
- Networks (optional)

After the cluster is successfully created, the following resources would be created in the cluster:
- Policies
- Admin Agents
- Admin workloads

Everything related to policy, service accounts, and service activation (and how these resources related to each other) would be hidden from the end user, simplifying their experience. 

## Additional Functional Requirements
In addition to addressing the top level user stories above, CNC should address the following functional requirements (gathered from interviews with a broad range of organizations that have failed to satisfactorily address their needs with other existing solutions):

1. CNC must be low friction to adopt and start using in any environment. Just install, and start defining compositions.
2. It should be pluggable - since building a new templating tool is a non-goal, and knowing that many organizations have some existing approach to this problem, users of other tools such as Helm, Kustomize, Crossplane etc. should be able to plug their existing templates into the CNC framework for simpler adoption in brownfield environments. 
3. It should enable monitoring capabilities so that a platform team can monitor each composition instance and the resources it manages and the relationships between resources. CNC should propagate errors from the underlying operators to the end-user.
4. It should support multi-tenant configurations, where different end users operate in different namespaces within the same cluster and may have permissions to use different compositions.
5. A composition should be able to reference fields in existing resources (which may or may not be defined by the composition).
6. Compositions should work consistently for any K8s resource, and be built as Kubernetes CRDs. 


## Risks and Mitigations
1. Building IDPs is complex and there are many edge cases. This will naturally drive CNC to become increasingly complex in its design. We should mitigate this by always optimizing for the “happy path”, ensuring that a new user can quickly start using CNC, while maintaining extensibility / pluggability where possible to account for edge cases and future use cases. 
2. CNC is intended to be specific to K8s, and not specific to any single vendor. We should be mindful of this and ensure that CNC features are generally applicable to the top level use case of “building an IDP using K8s”. Vendor-specific features should go in the operators that provide the mappings to vendor APIs (ACK, KCC, ASO etc.).
3. Many organizations have adopted some solution already, most commonly a set of Helm charts and custom hydration pipelines. Adopting new tools in a brownfield environment is challenging. We should make this easier where possible by allowing platform teams to “BYO templates” and use them within the CNC framework, so that they do not need to rewrite all their existing templates on Day 1. 

## Comparable Tools
There are existing tools that platform teams use to build their IDP. These tools, and how they compare with the goals of CNC, are described below.  
A number of these tools have client-side components. CNC explicitly intends to be server-side to provide operational consistency from a platform administration point of view, and to leverage K8s capabilities such as continuous reconciliation, RBAC, quotas, and namespace isolation out-of-the-box.
- Helm
- Crossplane
- KubeVela
- GitOps tools (Argo, Flux etc.)
- Cluster API (CAPI)
- Non-K8s-native tools (Terraform, Pulumi)

### CNC vs. Helm
Helm is a client-side package manager for Kubernetes that allows users to define, install, and upgrade Kubernetes applications as a set of YAML files (Helm charts). Helm is well suited for distribution of Kubernetes applications (releasing applications to end users), because it packages software as a unit. Helm is not intended for building IDPs, so does not have awareness of the resources defined within a chart and does not support executing logic to interpret platform user inputs to orchestrate cloud services, or monitoring of any resources created by the Helm chart.   

Since CNC is intended for building IDPs rather than distributing Kubernetes applications, a composition is not a “unit” in the way that a Helm chart is - it runs server-side, can interpret platform user inputs, orchestrate underlying K8s resources, control resource creation and deletion ordering, interact with resources defined outside of the composition, and propagate the status of defined resources back to the end user. It is a more dynamic system and less of a “black box” than a Helm chart, which cannot give visibility into the status of any resources created by the Helm chart. 

However, a basic composition that does not contain any logic, does not define a platform user interface, and does not reference external resources could look similar to a Helm chart - for this simpler use case either tool would suffice. Helm charts can also be used within the context of a larger service defined by a composition. For example, a composition may create a K8s cluster, and then install an nginx Helm chart in that cluster. 

### CNC vs. Crossplane
From the Crossplane documentation - “Crossplane connects your Kubernetes cluster to external, non-Kubernetes resources, and allows platform teams to build custom Kubernetes APIs to consume those resources.”

CNC is similar to Crossplane in the following ways:

1. Both projects aim to provide a framework for platform admins to define repeatable patterns through compositions (though the implementation of compositions is different). 
2. Both projects provide a way to expose a custom interface to platform users.
3. Both tools can leverage K8s continuous reconciliation to monitor and reconcile cloud resources. 

CNC differs from Crossplane in the following ways:

1. Crossplane introduces an abstraction layer (Crossplane Providers) for connecting to non-K8s services, whereas CNC does not introduce any abstractions and leverages K8s operators (such as KCC, ACK, and ASO) to provide connections to external cloud services as K8s CRDs. We believe this is a more flexible approach (since CNC compositions can work with any resource provided by any K8s operator), and more accessible to K8s users. CNC does not introduce any new providers, and the intention is that users will use the existing ecosystem of Kubernetes operators including Crossplane providers.

2. CNC takes a fully server-side-first approach, while Crossplane introduces a mix of server-side and client-side tooling. In Crossplane, compositions and platform user interfaces are defined as server-side “XRDs” (a Crossplane-specific K8s kind), and any logic required to interpret user inputs can be defined as a Crossplane Function, which is defined client-side and runs as an OCI image. CNC takes the approach of using K8s CRDs to define all compositions, platform user interface, and logic. This approach ensures that CNC can provide full visibility into the current and possible states of application infrastructure. 

3. CNC and Crossplane have different approaches to multi-tenancy. In Crossplane, all managed resources are exposed at the cluster scope, and the Crossplane installation is cluster scoped. One architecture that CNC aims to support is application teams being fully isolated by namespace, so CNC (and the K8s operators it uses such as ASO, KCC, and ACK) can be installed and used within a namespace. 

4. Crossplane is more opinionated in how compositions are defined, whereas CNC aims to provide flexibility in how compositions are defined. While CNC will provide a prescriptive “happy path”, platform teams can use the CNC framework with any development language / templating engine of their choice (for example, part of a composition could be defined by a Helm chart, and another part in CEL). CNC is specific to K8s, but platform admins can use the templating languages most familiar to them. In Crossplane, all compositions must be defined using the XRD structure. 

### CNC vs. KubeVela
The goal of KubeVela is to provide an abstraction layer for developers to deploy their applications against, similar to the goal of products like Cloud Foundry and Heroku. This differs from the goal of CNC, which is to help platform teams provide cloud services to their platform users in a consistent, secure and scalable manner. 

KubeVela allows developers to define their application deployment as a workflow. This workflow can leverage any CI/CD or GitOps tooling. For developers using KubeVela, their KubeVela deployment workflow could deploy an application that leverages CNC compositions as part of that application's definition. CNC does not focus on programmability for orchestrating deployments, and instead relies on the existing behavior of operators.

### CNC vs. GitOps Tools
The CNC framework can leverage templating languages such as CEL to build logic directly into a composition. For example, the logic to process platform user inputs can be defined in a composition, meaning all cloud resources and related logic exists as a CR in the cluster. This makes CNC a server-side DRY tool. This is helpful for platform teams building IDPs because the cluster becomes the one place to see both how cloud services are operating and what can be allowed to happen (and want cannot be allowed to happen, when combined with server-side policy tools like OPA Gatekeeper). 

CNC can be used with GitOps tools such as Argo CD and Flux CD in the following ways:
GitOps tooling can be used by platform admins to sync composition definitions from a git repo to the cluster. 
GitOps tooling can be used by platform users to create instances of a composition. The user inputs for the composition would be processed by CNC in the cluster after the GitOps tool has performed the sync.

### CNC vs. Cluster API (CAPI)
CNC can be used to create kubernetes clusters and surrounding cloud resources, which is also a goal of CAPI.  However, CAPI is a set of specialized operators for cluster lifecycle management, and CNC is a generic tool for orchestrating kubernetes resources.  For example, CNC cannot itself support node draining or kubernetes-aware version upgrade.  In contrast, CAPI cannot manage more than just kubernetes clusters.  A user that wants to orchestrate kubernetes clusters alongside additional resources might choose to use CNC to orchestrate CAPI resources with other CRDs. 

### CNC vs. Non-K8s-native Tools (Terraform, Pulumi)
The goal of CNC is to build repeatable K8s-native patterns, which is not a goal of non-K8s-native tools such as Terraform. Additionally, tools such as Terraform largely focus on infrastructure deployment (e.g. provisioning a K8s cluster), whereas CNC is more focused on defining services for end users such as developers or data scientists - these services may be a specific resource like a database, or a package that defines an entire application and it’s infrastructure. 
