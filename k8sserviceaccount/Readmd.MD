## Build image
go build -o clean clean.go
docker build -t clean:v0.1.0 . 

## Create object in k8s
kubectl create ns testns1
kubectl label ns testns1 env=autotest app=mysql
kubectl create ns testns2
kubectl label ns testns2 env=autotest app=mysql

kubectl create -f serviceaccount.yaml
kubectl logs clean -n loch

## Reference
[在Kubernetes Pod中使用Service Account访问API Server](https://tonybai.com/2017/03/03/access-api-server-from-a-pod-through-serviceaccount/)