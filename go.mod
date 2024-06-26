module github.com/leisheyoufu/golangstudy

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.0
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Shopify/sarama v1.26.1
	github.com/antlr4-go/antlr/v4 v4.13.0
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/coreos/bbolt v1.3.9 // indirect
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/emicklei/go-restful v2.9.6+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/garyburd/redigo v1.6.0
	github.com/go-mysql-org/go-mysql v1.3.0
	github.com/go-openapi/spec v0.19.7
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/consul/api v1.7.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/kr/pty v1.1.8
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/mozillazg/go-cos v0.13.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pingcap/check v0.0.0-20200212061837-5e12011dc712
	github.com/pingcap/dm v1.0.7
	github.com/pingcap/errors v0.11.5-0.20201126102027-b0a155152ca3
	github.com/pingcap/failpoint v0.0.0-20220801062533-2eaa32854a6c
	github.com/pingcap/parser v0.0.0-20210415081931-48e7f467fd74
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/common v0.14.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v0.0.0-20191125035519-b054a8dfd10d
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/tencentyun/cos-go-sdk-v5 v0.7.42
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/xitongsys/parquet-go v1.5.1
	github.com/xitongsys/parquet-go-source v0.0.0-20201108113611-f372b7d813be
	go.etcd.io/bbolt v1.3.9 // indirect
	go.uber.org/atomic v1.7.0
	golang.org/x/crypto v0.1.0
	golang.org/x/net v0.1.0
	golang.org/x/sync v0.5.0
	golang.org/x/text v0.4.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.27.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	gopkg.in/yaml.v2 v2.3.0
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

replace (
	github.com/coreos/bbolt v1.3.9 => go.etcd.io/bbolt v1.3.9
	go.etcd.io/bbolt v1.3.9 => github.com/coreos/bbolt v1.3.9
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
