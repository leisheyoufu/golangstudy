## Debug
dlv debug metadata/metadata.go -- --endpoints <brokers> --user admin --pass VyScFSogoxDl

go build -o produce producer/producer.go
./produce --endpoints <brokers> --user admin --pass 123456 --topic topic22


go build -o consume consumer/consumer.go
./consume --endpoints <brokers> --user golang22 --pass 123456 --topic topic22 --group group22


## Referebce
[获取group, topic, broker, partition各种指标参考](https://github.com/sundy-li/burrowx/blob/master/monitor/client.go)