## Start mysql service
```
docker network create service --subnet 172.18.0.0/16
docker run --name mysql-master -it --rm -e "MYSQL_ROOT_PASSWORD=123456" -p 3309:3306 --net=service -v `pwd`/conf/master:/etc/mysql/conf.d/ mysql:5.7
mysql -uroot -p123456 -h127.0.0.1 -P3309
```

## Start binlog syncer
```
o run main.go -port 3309
[2020/09/22 16:56:48] [info] binlogsyncer.go:144 create BinlogSyncer with config {1825 mysql 127.0.0.1 3309 root   utf8 false false <nil> false UTC false 0 0s 0s 0 false false 0}
[2020/09/22 16:56:48] [info] dump.go:199 skip dump, use last binlog replication pos (mysql-bin.000003, 194) or GTID set <nil>
[2020/09/22 16:56:48] [info] binlogsyncer.go:359 begin to sync binlog from position (mysql-bin.000003, 194)
[2020/09/22 16:56:48] [info] sync.go:25 start sync binlog at binlog file (mysql-bin.000003, 194)
[2020/09/22 16:56:48] [info] binlogsyncer.go:776 rotate to (mysql-bin.000003, 194)
[2020/09/22 16:56:48] [info] sync.go:68 received fake rotate event, next log name is mysql-bin.000003
```

## Execute sql inside another terminal
mysql -uroot -p123456 -h127.0.0.1 -P3309 -e "source user.sql;"

## Got the response
字符串是由canal注册的binlogHandler的OnRow方法打印
```
[2020/09/22 16:58:58] [info] binlogsyncer.go:776 rotate to (mysql-bin.000003, 194)
[2020/09/22 16:58:58] [info] sync.go:68 received fake rotate event, next log name is mysql-bin.000003
[2020/09/22 16:59:03] [info] sync.go:235 table structure changed, clear table cache: Test.User
User 1 is created with name Jack
User 1 name changed from Jack to Jonh
User 1 is deleted with name Jonh
```

OnRow返回的结构
```
// RowsEvent is the event for row replication.
type RowsEvent struct {
	Table  *schema.Table
	Action string
	// changed row list
	// binlog has three update event version, v0, v1 and v2.
	// for v1 and v2, the rows number must be even.
	// Two rows for one event, format is [before update row, after update row]
	// for update v0, only one row for a event, and we don't support this version.
	Rows [][]interface{}
	// Header can be used to inspect the event
	Header *replication.EventHeader
}
```
## Reference
[如何使用 Golang 处理 MySQL 的 binlog](https://studygolang.com/articles/21373)