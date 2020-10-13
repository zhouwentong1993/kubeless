module github.com/kubeless/redis-trigger

go 1.15

require (
	github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.10 // indirect
	github.com/coreos/prometheus-operator v0.0.0-20171201110357-197eb012d973 // indirect
	github.com/emicklei/go-restful v2.14.2+incompatible // indirect
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/kubeless/kubeless v1.0.0-alpha.8
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.0.0-20180103181712-d0becfa6529e // indirect
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v5.0.0+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200923155610-8b5066479488 // indirect
	k8s.io/utils v0.0.0-20201005171033-6301aaf42dc7 // indirect

)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20180103175015-389dfa299845
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20180103174757-bc110fd540ab
)
