module github.com/m88i/nexus-operator

go 1.14

require (
	github.com/RHsyseng/operator-utils v1.4.4
	// controller-runtime uses v0.1.0, klogv2 uses v0.2.0, which is the log module for k8s
	// as soon as they sync, we can migrate to v0.2.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.6
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.5.1
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
	github.com/m88i/aicura v0.2.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v0.0.0-20201005153912-821561a7f2a2
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6
	sigs.k8s.io/controller-runtime v0.6.3
)

replace (
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
	// latest klog whic uses logr 0.1.0
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.1.0
)
