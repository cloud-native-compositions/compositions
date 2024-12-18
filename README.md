# README

The is a README template for CNCF projects. Please start by renaming this to
README.md and deleting everything up to and including the "template begins here"
comment in this markdown file.

This is a template document for CNCF projects that requires editing
before it is ready to use. Read the markdown comments, `<!-- COMMENT -->`, for
additional guidance. The raw markdown uses `TODO` to identify areas that
require customization.  Replace [TODO: PROJECTNAME] with the name of your project.

<!-- template begins here-->

# Welcome to the Cloud Native Compositions Project!

<!-- Mission Statement -->
<!-- More information about crafting your mission statement with examples -->
<!-- https://contribute.cncf.io/maintainers/governance/charter/ -->

Cloud Native Compositions (CNC) is a K8s-native framework for building secure and reusable APIs for application teams to consume in a self-service manner. 

Many organizations are now choosing to build internal developer platforms (IDPs) on top of K8s. This has the benefit of enabling a platform team to provide self-service infrastructure to platform users while leveraging the K8s ecosystem. However, this approach also introduces challenges - platform teams need to define and implement the contract for the services that are exposed to end users (across many vendor APIs), monitor and control service access and consumption, and attempt to simplify the platform user experience for non-experts in K8s or cloud provider APIs. 

Cloud Native Compositions solves for the challenge of how platform teams define services for their K8s-based IDP. It is a framework to define repeatable patterns, called compositions, for how cloud services should be consumed in their organization. Any Kubernetes user can benefit from these K8s-native patterns to reduce context switching even if they are not building a platform, although we are optimizing for the platform use-case.


[TODO:
Implementation, strategy and architecture].

<!-- If CNCF:
Cloud Native Compositions is hosted by the [Cloud Native Computing Foundation (CNCF)](https://cncf.io).
-->

## Getting Started

<!-- Include enough details to get started using, or at least building, the
project here and link to other docs with more detail as needed.  Depending on
the nature of the project and its current development status, this might
include:
* quick installation/build instructions
* a few simple examples of use
* basic prerequisites
--> 

## Contributing
<!-- Template: https://github.com/cncf/project-template/blob/main/CONTRIBUTING.md -->

Our project welcomes contributions from any member of our community. To get
started contributing, please see our [Contributor Guide](TODO: Link to
CONTRIBUTING.md).

## Scope
<!-- If this section is too long, you might consider moving it to a SCOPE.md -->
<!-- More information about creating your scope with links to examples -->
<!-- https://contribute.cncf.io/maintainers/governance/charter/ -->

### In Scope

Cloud Native Compositions is intended to empower platform teams to build K8s-native internal developer platforms that: 

- Abstract and resolve complex interdependencies between resources, enabling end user self-service for platform teams and simplified resource management for small dev teams
- Provide visibility into how an application is interacting with cloud services
- Are easier and cheaper to build and maintain than existing solutions
- Are compatible with any K8s operator, and existing packaging and templating tools
- Leverage existing extensibility types (CRDs) as the API for interacting with cloud resources

### Out of Scope

Cloud Native Compositions will be used in a cloud native environment with other
tools. The following specific functionality will therefore not be incorporated:

- A new templating tool. The framework should be pluggable, such that users can BYO existing templating tools such as CEL, Helm, Jinja etc. 
- All of the tooling needed by platform users / downstream users. CNC will help platform administrators provide services to platform users. The customers of CNC are platform administrators. 
- New types of extensibility for Kubernetes besides CRDs.



Cloud Native Compositions implements [TODO: List of major features, existing or
planned], through [TODO: Implementation
requirements/language/architecture/etc.]. It will not cover [TODO: short list
of excluded items]

## Communications

<!-- Fill in the communications channels you actually use.  These should all be public channels anyone
can join, and there should be several ways that users and contributors can reach project maintainers. 
If you have recurring/regular meetings, list those or a link to a publicy-readable calendar so that
prospective contributors know when and where to engage with you. -->

[TODO: Details (with links) to meetings, mailing lists, Slack, and any other communication channels]

* User Mailing List:
* Developer Mailing List:
* Slack Channel:
* Public Meeting Schedule and Links: 
* Social Media:
* Other Channel(s), If Any:

## Resources
- [CNC project proposal](cnc-proposal.md)

[TODO: Add links to other helpful information (roadmap, docs, website, etc.)]

## License

This project is licensed under the [Apache License 2.0](LICENSE)

## Conduct

We follow the [CNCF Code of Conduct](CODE_OF_CONDUCT.md).
