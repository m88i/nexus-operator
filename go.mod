module github.com/m88i/nexus-operator

require (
	github.com/RHsyseng/operator-utils v0.0.0-20191024171429-aada4a416142
	github.com/go-openapi/spec v0.19.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.11.0
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208
	sigs.k8s.io/controller-runtime v0.2.0
)

replace (
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190814194116-a94e914914f4+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190813201236-5a5508328169
)

// Pinned to kubernetes-1.14.1
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.11.0
