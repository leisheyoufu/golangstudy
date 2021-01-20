package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/leisheyoufu/golangstudy/mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

var host = flag.String("host", "127.0.0.1", "MySQL host")
var port = flag.Int("port", 3306, "MySQL port")
var user = flag.String("user", "root", "MySQL user, must have replication privilege")
var password = flag.String("password", "123456", "MySQL password")

type MyEventHandler struct {
	canal.DummyEventHandler
}

//监听数据记录
func (h *MyEventHandler) OnRow(ev *canal.RowsEvent) error {
	//record := fmt.Sprintf("%s %v %v %v %s\n",e.Action,e.Rows,e.Header,e.Table,e.String())

	//库名，表名，行为，数据记录
	record := fmt.Sprintf("%v %v %s %v\n", ev.Table.Schema, ev.Table.Name, ev.Action, ev.Rows)
	fmt.Printf("Binlog: OnRow Begin %s Rows=%d\n", record, len(ev.Rows))

	//此处是参考 https://github.com/gitstliu/MysqlToAll 里面的获取字段和值的方法
	for i, _ := range ev.Rows {
		var row string
		for j, currColumn := range ev.Table.Columns {
			//字段名，字段的索引顺序，字段对应的值
			// len(ev.Rows) 一次有多行
			fmt.Println(ev.Rows[i])
			row += fmt.Sprintf("%v %v\t data columns=%d\t", currColumn.Name, ev.Rows[i][j], len(ev.Rows[i]))
		}
		fmt.Printf("Binlog: OnRow columes=%d info:%s\n", len(ev.Table.Columns), row)
	}

	fmt.Printf("Binlog: OnRow end\n")

	return nil
}

//创建、更改、重命名或删除表时触发，通常会需要清除与表相关的数据，如缓存。It will be called before OnDDL.
func (h *MyEventHandler) OnTableChanged(schema string, table string) error {
	//库，表
	record := fmt.Sprintf("%s %s \n", schema, table)
	fmt.Printf("Binlog: OnTable Changed\n", record)
	return nil
}

//监听binlog日志的变化文件与记录的位置
func (h *MyEventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	//源码：当force为true，立即同步位置
	record := fmt.Sprintf("%v %v \n", pos.Name, pos.Pos)
	fmt.Printf("Binlog: OnPosSynced\n", record)
	return nil
}

//当产生新的binlog日志后触发(在达到内存的使用限制后（默认为 1GB），会开启另一个文件，每个新文件的名称后都会有一个增量。)
func (h *MyEventHandler) OnRotate(r *replication.RotateEvent) error {
	//record := fmt.Sprintf("On Rotate: %v \n",&mysql.Position{Name: string(r.NextLogName), Pos: uint32(r.Position)})
	//binlog的记录位置，新binlog的文件名
	record := fmt.Sprintf("On Rotate %v %v \n", r.Position, r.NextLogName)
	fmt.Printf("Binlog: %s", record)
	return nil

}

// Begin transation
func (h *MyEventHandler) OnBegin(pos mysql.Position, eh *replication.EventHeader) error {
	fmt.Printf("Binlog: OnBegin\n")
	return nil
}

// includes begin, commit, query
func (h *MyEventHandler) OnQuery(eh *replication.EventHeader, e *replication.QueryEvent) error {
	fmt.Printf("Binlog: OnQuery %s\n", string(e.Query))
	return nil
}

// create alter drop truncate(删除当前表再新建一个一模一样的表结构)
func (h *MyEventHandler) OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	//binlog日志的变化文件与记录的位置
	record := fmt.Sprintf("%v %v\n", nextPos.Name, nextPos.Pos)
	query_event := fmt.Sprintf("%v\n %v\n %v\n %v\n %v\n",
		queryEvent.ExecutionTime,         //猜是执行时间，但测试显示0
		string(queryEvent.Schema),        //库名
		string(queryEvent.Query),         //变更的sql语句
		string(queryEvent.StatusVars[:]), //测试显示乱码
		queryEvent.SlaveProxyID)          //从库代理ID？
	fmt.Println("Binlog: OnDDL start:", record, query_event)
	fmt.Println("Binlog: OnDDL end\n")
	return nil
}

// commit
func (h *MyEventHandler) OnXID(nextPos mysql.Position) error {
	fmt.Printf("Binlog: OnXID commit\n")
	return nil
}

func (h *MyEventHandler) OnMariaDBGTID(eh *replication.EventHeader, e *replication.MariadbGTIDEvent) error {
	gtid, err := mysql.ParseMariadbGTIDSet(e.GTID.String())
	if err != nil {
		fmt.Printf("Parse gtid error %v\n", err)
		return err
	}
	fmt.Printf("Binlog: gtid=%s\n", gtid.String())
	return nil
}

func (h *MyEventHandler) OnGTID(gtid mysql.GTIDSet) error {
	fmt.Printf("Binlog: gtid=%s:%d\n", gtid.String())
	return nil
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}

func main() {
	flag.Parse()
	//读取toml文件格式
	//canal.NewConfigWithFile()
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d", *host, *port)
	cfg.User = *user
	cfg.Password = *password

	cfg.Dump.TableDB = ""
	cfg.Dump.ExecutionPath = ""
	fmt.Println(cfg)
	c, err := canal.NewCanal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	c.SetEventHandler(&MyEventHandler{})
	//mysql-bin.000004, 1027
	pos, err := c.GetMasterPos()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	//从头开始监听
	//c.Run()
	//根据位置监听
	c.RunFrom(pos)
}
