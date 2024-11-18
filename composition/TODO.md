# Code Organization

1. Move expander composition/expanders/jinja2 -> expanders/jinja2 
2. Move expander composition/expanders/getter -> expanders/getter 
3. Remove `inline` and Job based expander code and dockerfile, make targets
4. move away from using github.com/wzshiming/easycel
5. Add fields to composition/config/samples/*
6. kube-rbac-proxy is not needed anymore. remove it from manifests.  ASO PR: Azure/azure-service-operator#3833 . controller-runtime feature: kubernetes-sigs/controller-runtime#2407, they're dropping it in the scaffolding too here: kubernetes-sigs/kubebuilder#3871
7. composition/config/manager/manager.yaml add readOnlyRootFilesystem
8. expanderConfig.configRef.namespace => force this to be the same ns as composition. so remove this ? 
9. Should compositions be namespace scoped ?