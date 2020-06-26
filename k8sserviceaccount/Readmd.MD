## Build image
go build main.go
docker build -t clscu/serviceaccount:latest . 

## Create object in k8s
kubectl create -f serviceaccount.yaml
kubectl create -f deployment.yaml
kubectl logs serviceaccountexample -n loch

## Reference
[在Kubernetes Pod中使用Service Account访问API Server](https://tonybai.com/2017/03/03/access-api-server-from-a-pod-through-serviceaccount/)