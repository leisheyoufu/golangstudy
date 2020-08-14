module github.com/leisheyoufu/golangstudy

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.0
)

require (
	github.com/Shopify/sarama v1.26.1
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170926063155-7524189396c6
	github.com/garyburd/redigo v1.6.0
	github.com/go-openapi/spec v0.19.7
	github.com/golang/protobuf v1.3.4
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kr/pty v1.1.8
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pingcap/dm v1.0.6
	github.com/pingcap/failpoint v0.0.0-20200702092429-9f69995143ce
	github.com/siddontang/go-mysql v0.0.0-20200529051020-8250ec49cbbf
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/objx v0.2.0 // indirect
	golang.org/x/crypto v0.0.0-20200204104054-c9f3fb736b72
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.2
	k8s.io/kubernetes v1.13.1
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)
