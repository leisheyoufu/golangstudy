package client

import (
	"github.com/leisheyoufu/golangstudy/redis/common"
)

const (
	Address = "192.168.126.10:6379"
)

var (
	log *common.Logger
)

func init() {
	common.InitLogger()
	log = common.GetLogger("/client")

}
