package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/leisheyoufu/golangstudy/mysql/canal"
	"github.com/leisheyoufu/golangstudy/mysql/replication"
)

var host = flag.String("host", "127.0.0.1", "MySQL host")
var port = flag.Int("port", 3306, "MySQL port")
var user = flag.String("user", "root", "MySQL user, must have replication privilege")
var password = flag.String("password", "123456", "MySQL password")
var startFile = flag.String("file", "", "mysql binlog file")
var startPos = flag.Uint("pos", 0, "start binlog pos")
var endFile = flag.String("endfile", "binlog.999999", "end binlog pos")
var endPos = flag.Uint("endpos", 99999, "end binlog pos")
var gtid = flag.String("gtid", "", "mysql gtid")
var flavor = flag.String("flavor", mysql.MySQLFlavor, "mysql or mariadb")
var db = flag.String("db", "", "db to sync")
var table = flag.String("table", "", "table to sync")
var value = flag.String("value", "", "data to search")
var limit = flag.Uint("limit", 1000000, "limit count")

type MyEventHandler struct {
	canal.DummyEventHandler
	gtid  string
	pos   mysql.Position
	db    string
	table string
	value string
	count int64
}

func Equal(a interface{}, b string) bool {
	var val string
	switch value := a.(type) {
	case int:
		val = strconv.Itoa(value)
	case int16:
		val = strconv.FormatInt(int64(value), 10)
	case int32:
		val = strconv.FormatInt(int64(value), 10)
	case int64:
		val = strconv.FormatInt(value, 10)
	case int8:
		val = strconv.FormatInt(int64(value), 10)
	case uint8:
		val = strconv.FormatUint(uint64(value), 10)
	case uint16:
		val = strconv.FormatUint(uint64(value), 10)
	case uint32:
		val = strconv.FormatUint(uint64(value), 10)
	case uint64:
		val = strconv.FormatUint(value, 10)
	case string:
		val = value
	}
	if val == b {
		return true
	}

	return false
}

func PosLessOrEqual(pos1 mysql.Position, pos2 mysql.Position) bool {
	parts1 := strings.Split(pos1.Name, ".")
	index1, _ := strconv.ParseUint(parts1[len(parts1)-1], 10, 64)

	parts2 := strings.Split(pos2.Name, ".")
	index2, _ := strconv.ParseUint(parts2[len(parts2)-1], 10, 64)

	if index1 < index2 {
		return true
	}
	if index1 == index2 && pos1.Pos < pos2.Pos {
		return true
	}
	return false
}

// 监听数据记录
func (h *MyEventHandler) OnRow(ev *canal.RowsEvent) error {
	h.count++
	found := true
	printAll := false
	if h.value != "" {
		found = false
	}

	if h.db == "" || h.table == "" {
		printAll = true
	}
	if !printAll {
		if !(h.db == ev.Table.Schema && h.table == ev.Table.Name) {
			return nil
		}
	}
	targetIndexs := make([]int, 0)
	//fmt.Printf("Row event %s %s.%s Rows=%d\n", ev.Action, ev.Table.Schema, ev.Table.Name, len(ev.Rows))

	//此处是参考 https://github.com/gitstliu/MysqlToAll 里面的获取字段和值的方法
	if !found {
		for i, _ := range ev.Rows {
			for j, _ := range ev.Table.Columns {
				//if col.Name == "id" {
				//	fmt.Printf("cltest column id val %v\n", ev.Rows[i][j])
				//}
				//if col.Name == "id" {
				//	if ev.Rows[i][j].(int32) == 29906751 {
				//		found = true
				//	}
				//	//fmt.Printf("cltest id column type %v", reflect.TypeOf(ev.Rows[i][j]))
				//}
				if Equal(ev.Rows[i][j], h.value) {
					targetIndexs = append(targetIndexs, i)
					found = true
				}
			}
		}
	}
	if printAll {
		for i, _ := range ev.Rows {
			if ev.Action == "update" && i == 0 {
				fmt.Printf("Before:\n")
			}
			if ev.Action == "update" && i == 1 {
				fmt.Printf("After:\n")
			}
			for j, currColumn := range ev.Table.Columns {
				fmt.Printf("db %s table %s column %s = %v\n", ev.Table.Schema, ev.Table.Name, currColumn.Name, ev.Rows[i][j])
			}
		}
	}
	if found {
		for _, targetIndex := range targetIndexs {
			for j, currColumn := range ev.Table.Columns {
				fmt.Printf("db %s table %s column %s = %v\n", ev.Table.Schema, ev.Table.Name, currColumn.Name, ev.Rows[targetIndex][j])
				if currColumn.Name == "remark" {
					fmt.Printf("cltest remark json %v\n", string(ev.Rows[targetIndex][j].([]byte)))
				}
			}
		}

	}
	end := mysql.Position{Pos: uint32(*endPos), Name: *endFile}
	if !PosLessOrEqual(h.pos, end) {
		fmt.Printf("end pos %d reached\n", *endPos)
		os.Exit(0)
	}
	//fmt.Printf("Binlog: OnRow end\n")
	return nil
}

// 创建、更改、重命名或删除表时触发，通常会需要清除与表相关的数据，如缓存。It will be called before OnDDL.
func (h *MyEventHandler) OnTableChanged(schema string, table string) error {
	//库，表
	//record := fmt.Sprintf("%s %s \n", schema, table)
	if !(h.db == "" && h.table == "" || h.db == schema && h.table == table) {
		return nil
	}
	if h.count > int64(*limit) {
		return nil
	}
	fmt.Printf("Binlog: OnTable Changed %s.%s\n", schema, table)

	return nil
}

// 监听binlog日志的变化文件与记录的位置
func (h *MyEventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	//源码：当force为true，立即同步位置
	h.pos = mysql.Position{
		Name: pos.Name,
		Pos:  pos.Pos,
	}
	//fmt.Printf("Sync pos to %v\n", h.pos)
	return nil
}

// 当产生新的binlog日志后触发(在达到内存的使用限制后（默认为 1GB），会开启另一个文件，每个新文件的名称后都会有一个增量。)
func (h *MyEventHandler) OnRotate(r *replication.RotateEvent) error {
	//record := fmt.Sprintf("On Rotate: %v \n",&mysql.Position{Name: string(r.NextLogName), Pos: uint32(r.Position)})
	//binlog的记录位置，新binlog的文件名
	h.pos = mysql.Position{
		Name: string(r.NextLogName),
		Pos:  uint32(r.Position),
	}
	fmt.Printf("Rotate pos to %v\n", h.pos)
	return nil

}

// Begin transation
func (h *MyEventHandler) OnBegin(pos mysql.Position, eh *replication.EventHeader) error {
	h.count++
	h.pos = mysql.Position{
		Name: pos.Name,
		Pos:  pos.Pos,
	}
	return nil
}

// includes begin, commit, query
func (h *MyEventHandler) OnQuery(nextPos mysql.Position, e *replication.QueryEvent) error {
	h.count++
	//if !(h.db == "" && h.table == "" || h.db == string(e.Schema) && strings.Contains(string(e.Query), h.table)) {
	//	return nil
	//}
	//if !(h.db == string(e.Schema)){
	//	return nil
	//}
	//if h.count > int64(*limit) {
	//	return nil
	//}
	if !(strings.Contains(string(e.Query), h.db)) {
		return nil
	}
	fmt.Printf("Binlog: OnQuery gtid=%s pos=%v db=%s %s\n", h.gtid, h.pos, string(e.Schema), string(e.Query))
	h.pos = nextPos
	end := mysql.Position{Pos: uint32(*endPos), Name: *endFile}
	if !PosLessOrEqual(h.pos, end) {
		fmt.Printf("end pos %d reached\n", *endPos)
		os.Exit(0)
	}
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
	if h.count > int64(*limit) {
		return nil
	}
	fmt.Println("Binlog: OnDDL start:", record, query_event)
	fmt.Println("Binlog: OnDDL end\n")
	return nil
}

// commit
func (h *MyEventHandler) OnXID(nextPos mysql.Position) error {
	h.count++
	if h.count > int64(*limit) {
		return nil
	}
	//fmt.Printf("At pos %v, gtid %s OnXID\n", h.pos, h.gtid)
	h.pos = mysql.Position{
		Name: nextPos.Name,
		Pos:  nextPos.Pos,
	}
	return nil
}

func (h *MyEventHandler) OnMariaDBGTID(eh *replication.EventHeader, e *replication.MariadbGTIDEvent) error {
	h.count++
	_, err := mysql.ParseMariadbGTIDSet(e.GTID.String())
	if err != nil {
		fmt.Printf("Parse gtid error %v\n", err)
		return err
	}
	h.gtid = e.GTID.String()
	return nil
}

func (h *MyEventHandler) OnGTID(gtid mysql.GTIDSet) error {
	h.gtid = gtid.String()
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
	//cfg.UseDecimal = false

	cfg.Dump.TableDB = ""
	cfg.Dump.ExecutionPath = ""
	fmt.Println(cfg)
	c, err := canal.NewCanal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	handler := &MyEventHandler{db: *db, table: *table, value: *value}
	c.SetEventHandler(handler)
	//mysql-bin.000004, 1027
	var pos mysql.Position
	if *startFile != "" {
		pos = mysql.Position{Name: *startFile, Pos: uint32(*startPos)}
	} else {
		pos, err = c.GetMasterPos()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
	}
	go func() {
		for {
			fmt.Printf("scaned pos %s gtid %s\n", handler.pos.String(), handler.gtid)
			time.Sleep(5 * time.Second)
		}

	}()

	handler.pos = pos
	if *gtid != "" {
		startGtid, err := mysql.ParseGTIDSet(*flavor, *gtid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse gtid %s\n", err.Error())
			os.Exit(1)
		}
		err = c.StartFromGTID(startGtid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start from gtid %v, %s\n", startGtid, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	} else if *startFile != "" && *startPos != 0 {
		pos = mysql.Position{
			Name: *startFile,
			Pos:  uint32(*startPos),
		}
	}
	err = c.RunFrom(pos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start from pos %v, %s\n", pos, err.Error())
		os.Exit(1)
	}
}
