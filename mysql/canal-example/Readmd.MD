## Prepare data (deprecate)
mysql> create database sbtest;
### Mac
brew install sysbench
// prepare data in sbtest database
sysbench /usr/local/Cellar/sysbench/1.0.20/share/sysbench/oltp_delete.lua --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password='123456' --mysql-db=sbtest --db-driver=mysql --tables=1 --table-size=1 prepare

sysbench /usr/local/Cellar/sysbench/1.0.20/share/sysbench/oltp_delete.lua --mysql-host=127.0.0.1 --mysql-port=3306 --mysql-user=root --mysql-password='123456' --mysql-db=sbtest --db-driver=mysql --tables=3 --table-size=100 cleanup

```
mysql> show master status;
+---------------+----------+--------------+------------------+-------------------+
| File          | Position | Binlog_Do_DB | Binlog_Ignore_DB | Executed_Gtid_Set |
+---------------+----------+--------------+------------------+-------------------+
| binlog.000004 |  5730584 |              |                  |                   |
+---------------+----------+--------------+------------------+-------------------+

mysqlbinlog --no-defaults --read-from-remote-server --host=127.0.0.1  --user=root --password=New123 --port 3319 -v --base64-output=decode-rows --start-position=3579 mysql-bin.000003
```
go run main.go
```
[2020/06/22 11:14:28] [info] sync.go:235 table structure changed, clear table cache: sbtest.sbtest3
sbtest sbtest3 

OnDDL: binlog.000004 5791058
 0
 sbtest
 DROP TABLE IF EXISTS `sbtest3` /* generated by server */
  �Estd--�
          sbtest:�
 17
```

## Start mysql service
```
docker network create service --subnet 172.18.0.0/16
docker run --name mysql-master -d --rm -e "MYSQL_ROOT_PASSWORD=123456" -p 3309:3306 --net=service -v `pwd`/conf/master:/etc/mysql/conf.d/ mysql:5.7
mysql -uroot -p123456 -h127.0.0.1 -P3309
```

## Run
go run main.go --port 3319
mysql -uroot -p123456 -h127.0.0.1 -P3309 -e "drop database Test;"
mysql -uroot -p123456 -h127.0.0.1 -P3309 -e "source user.sql;"
dlv debug main.go -- --port 3309
(dlv) b main.main
(dlv) sources


## Result
```
[2020/09/22 20:09:08] [info] sync.go:235 table structure changed, clear table cache: Test.User
Test User 

OnDDL: mysql-bin.000003 2260
 0
 Test
 create table Test.User
(
  id int auto_increment primary key,
  name varchar(40) null,
  status enum("active","deleted") DEFAULT "active",
  created timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
) engine=InnoDB
  �Vstd--
        Test
 12

OnPosSynced mysql-bin.000003 2260 

Test User insert [[1 Jack 1 2020-09-22 20:09:08]]

row info: id 0 1

row info: name 1 Jack

row info: status 2 1

row info: created 3 2020-09-22 20:09:08

OnPosSynced mysql-bin.000003 2533 

Test User update [[1 Jack 1 2020-09-22 20:09:08] [1 Jonh 1 2020-09-22 20:09:08]]

row info: id 0 1

row info: name 1 Jonh

row info: status 2 1

row info: created 3 2020-09-22 20:09:08

OnPosSynced mysql-bin.000003 2822 

Test User delete [[1 Jonh 1 2020-09-22 20:09:08]]

row info: id 0 1

row info: name 1 Jonh

row info: status 2 1

row info: created 3 2020-09-22 20:09:08

OnPosSynced mysql-bin.000003 3095 
```

## Debug
dlv debug main.go
### Qeury Event
```
(dlv) p e
*github.com/siddontang/go-mysql/replication.QueryEvent {
        SlaveProxyID: 4,
        ExecutionTime: 0,
        ErrorCode: 0,
        StatusVars: []uint8 len: 33, cap: 79, [0,0,0,0,0,1,32,0,128,86,0,0,0,0,6,3,115,116,100,4,45,0,45,0,8,0,12,1,84,101,115,116,0],
        Schema: []uint8 len: 4, cap: 46, [84,101,115,116],
        Query: []uint8 len: 34, cap: 41, [99,114,101,97,116,101,32,100,97,116,97,98,97,115,101,32,105,102,32,110,111,116,32,101,120,105,115,116,115,32,84,101,115,116],
        GSet: github.com/siddontang/go-mysql/mysql.GTIDSet nil,}

(dlv) p string(e.Query)
"create database if not exists Test"
(dlv) p string(e.Schema)
"Test"
(dlv) p stmts
[]github.com/pingcap/parser/ast.StmtNode len: 1, cap: 1, [
        *github.com/pingcap/parser/ast.CreateDatabaseStmt {
                ddlNode: (*"github.com/pingcap/parser/ast.ddlNode")(0xc000080bc0),
                IfNotExists: true,
                Name: "Test",
                Options: []*github.com/pingcap/parser/ast.DatabaseOption len: 0, cap: 0, [],},

(dlv) p string(e.Query)
"create table Test.User\n(\n  id int auto_increment primary key,\n  "
(dlv) p string(e.Query)
"BEGIN"

### One insert statement contains:
- *github.com/siddontang/go-mysql/replication.GTIDEvent
- QueryEvent (BEGIN)
- github.com/siddontang/go-mysql/replication.Event(*github.com/siddontang/go-mysql/replication.TableMapEvent)
  (dlv) p string(e.Table)
   "User"
- *github.com/siddontang/go-mysql/replication.RowsEvent
- *github.com/siddontang/go-mysql/replication.XIDEvent {
        XID: 31,
        GSet: github.com/siddontang/go-mysql/mysql.GTIDSet nil,}

## 一次插入操作的event(innodb)
```
client event *replication.GTIDEvent
client event *replication.QueryEvent
cltest: QueryEvent BEGIN
client event *replication.TableMapEvent
client event *replication.RowsEvent  // rows Event包含多行的记录
online_ddl t2 insert [[1 20001 José Arcadio Buendía <nil>] [2 20002 Úrsula Iguarán <nil>] [3 20003 José Arcadio <nil>]]

row info: id 0 3

row info: uid 1 20003

row info: name 2 José Arcadio

row info: info 3 <nil>

client event *replication.XIDEvent
OnPosSynced mysql-bin.000003 5995

## 一次插入操作的event(innodb)
```
client event *replication.GTIDEvent
client event *replication.QueryEvent
cltest: QueryEvent BEGIN
client event *replication.TableMapEvent
client event *replication.RowsEvent
client event *replication.QueryEvent
cltest: QueryEvent COMMIT
```

## Reference
[利用基于canal的go库来监听binlog日志](https://blog.csdn.net/Marcus929/article/details/103871699)
[Cannot switch database with USE command](https://github.com/go-sql-driver/mysql/issues/173)