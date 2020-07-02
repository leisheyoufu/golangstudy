package main

import (
	"fmt"

	"github.com/pingcap/dm/pkg/binlog/event"
	"github.com/pingcap/dm/pkg/gtid"
	gmysql "github.com/siddontang/go-mysql/mysql"
)

func main() {
	var (
		flavor          = gmysql.MySQLFlavor
		serverID uint32 = 101
		gSetStr         = "3ccc475b-2343-11e7-be21-6c0b84d59f30:1-14,406a3f61-690d-11e7-87c5-6c92bf46f384:1-94321383,53bfca22-690d-11e7-8a62-18ded7a37b78:1-495,686e1ab6-c47e-11e7-a42c-6c92bf46f384:1-34981190,03fc0263-28c7-11e7-a653-6c0b84d59f30:1-7041423,05474d3c-28c7-11e7-8352-203db246dd3d:1-170,10b039fc-c843-11e7-8f6a-1866daf8d810:1-308290454"
		gSet     gtid.Set
	)
	gSet, err := gtid.ParserGTID(flavor, gSetStr)
	if err != nil {
		fmt.Printf("Parse GTID failed +%v\n", err)
	}

	events, data, _ := event.GenCommonFileHeader(flavor, serverID, gSet)

	fmt.Println(events)
	fmt.Println(string(data))

}
