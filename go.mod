module github.com/m88i/nexus-operator

require (
	github.com/RHsyseng/operator-utils v0.0.0-20200709142328-d5a5812a443f
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1
	github.com/heroku/docker-registry-client v0.0.0-20190909225348-afc9e1acc3d5
	github.com/m88i/aicura v0.2.0
	github.com/openshift/api v0.0.0-20200312145924-779a333deb13 // branch release-3.11
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
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.3 // Required by prometheus-operator
)

go 1.14
