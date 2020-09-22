package pkg

import (
	"fmt"
	"runtime/debug"

	"github.com/leisheyoufu/golangstudy/mysql/canal"
)

type binlogHandler struct {
	canal.DummyEventHandler
	BinlogParser
}

func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Print(r, " ", string(debug.Stack()))
		}
	}()

	// base value for canal.DeleteAction or canal.InsertAction
	var n = 0
	var k = 1

	if e.Action == canal.UpdateAction {
		n = 1
		k = 2
	}

	for i := n; i < len(e.Rows); i += k {

		key := e.Table.Schema + "." + e.Table.Name

		switch key {
		case User{}.SchemaName() + "." + User{}.TableName():
			user := User{}
			h.GetBinLogData(&user, e, i)
			switch e.Action {
			case canal.UpdateAction:
				oldUser := User{}
				h.GetBinLogData(&oldUser, e, i-1)
				fmt.Printf("User %d name changed from %s to %s\n", user.Id, oldUser.Name, user.Name)
			case canal.InsertAction:
				fmt.Printf("User %d is created with name %s\n", user.Id, user.Name)
			case canal.DeleteAction:
				fmt.Printf("User %d is deleted with name %s\n", user.Id, user.Name)
			default:
				fmt.Printf("Unknown action")
			}
		}

	}
	return nil
}

func (h *binlogHandler) String() string {
	return "binlogHandler"
}

func BinlogListener(endpoint string, user string, password string) {
	c, err := getDefaultCanal(endpoint, user, password)
	if err == nil {
		coords, err := c.GetMasterPos()
		if err == nil {
			c.SetEventHandler(&binlogHandler{})
			c.RunFrom(coords)
		}
	}
}

func getDefaultCanal(endpoint string, user string, password string) (*canal.Canal, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = endpoint
	cfg.User = user
	cfg.Password = password
	cfg.Flavor = "mysql"

	cfg.Dump.ExecutionPath = ""

	return canal.NewCanal(cfg)
}
