module github.com/leisheyoufu/golangstudy

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.0
)

require (
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/garyburd/redigo v1.6.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kr/pty v1.1.8
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20191128160524-b544559bb6d1
	golang.org/x/net v0.0.0-20191126235420-ef20fe5d7933
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/grpc v1.25.1
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)
