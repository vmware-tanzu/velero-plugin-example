module github.com/vmware-tanzu/velero-plugin-example

go 1.14

require (
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/hashicorp/go-hclog v0.9.2 // indirect
	github.com/hashicorp/go-plugin v1.0.1-0.20190610192547-a1bc61569a26 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/vmware-tanzu/velero v1.6.2
	k8s.io/api v0.19.12
	k8s.io/apimachinery v0.19.12
)

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
