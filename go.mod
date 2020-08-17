module github.com/m88i/nexus-operator

require (
	github.com/RHsyseng/operator-utils v0.0.0-20200709142328-d5a5812a443f
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1
	github.com/m88i/aicura v0.2.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/zap v1.15.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190814194116-a94e914914f4+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190813201236-5a5508328169
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM

replace k8s.io/client-go => k8s.io/client-go v0.18.3 // Required by prometheus-operator

go 1.14
