## Start consul service
docker run --rm --name consul -p 8500:8500 bitnami/consul:latest

注册服务： curl -X PUT -d '{"id": "job1","name": "job1","address": "192.168.56.12","port": 9100,"tags": ["service"],"checks": [{"http": "http://192.168.56.12:9100/","interval": "5s"}]}' http://127.0.0.1:8500/v1/agent/service/register
查询所有服务： curl http://127.0.0.1:8500/v1/catalog/services
注销服务：curl --request PUT http://127.0.0.1:8500/v1/agent/service/deregister/job4

## Build and Run
```
go run main.go
Curret consul services:job3 job4 job5 
Curret consul services:job4 job5
```
## Reference
[golang使用服务发现系统consul](https://studygolang.com/articles/9980)
[consul api service](https://www.consul.io/api-docs/agent/service)