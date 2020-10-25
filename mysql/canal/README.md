### 检查是否排除在外的表，排除sync就不解析binlog语句
```
key := fmt.Sprintf("%s.%s", db, table)
// if table is excluded, return error and skip parsing event or dump
if !c.checkTableMatch(key) {
    return nil, ErrExcludedTable
}
```

### 作为slave端，开始接收binlog event stream
```
// StartSync starts syncing from the `pos` position.
func (b *BinlogSyncer) StartSync(pos Position) (*BinlogStreamer, error) {
	log.Infof("begin to sync binlog from position %s", pos)

	b.m.Lock()
	defer b.m.Unlock()

	if b.running {
		return nil, errors.Trace(errSyncRunning)
	}

	if err := b.prepareSyncPos(pos); err != nil {
		return nil, errors.Trace(err)
	}

	return b.startDumpStream(), nil
}
```
### b.startDumpStream 会构建了下面的BinlogStreamer, 开启channel接收BinlogEvent
```
func newBinlogStreamer() *BinlogStreamer {
	s := new(BinlogStreamer)

	s.ch = make(chan *BinlogEvent, 10240)
	s.ech = make(chan error, 4)

	return s
}
```
### BinlogStreamer新建一个协程处理event
```
func (b *BinlogSyncer) onStream(s *BinlogStreamer) {
    ....
    if err = b.parseEvent(s, data); err != nil {  // 会检查收到的event包是否正确，然后发到s.ch(BinlogEvent channel)
```
### 会对event type 进行解析，
解析binlog的关键函数 runSyncBinlog, 这里Event是在parseEvent中根据二进制封装后生成的
```
func (c *Canal) runSyncBinlog() error {
	s, err := c.startSyncer()
	if err != nil {
		return err
	}

	savePos := false
	force := false

	// The name of the binlog file received in the fake rotate event.
	// It must be preserved until the new position is saved.
	fakeRotateLogName := ""

	for {
		ev, err := s.GetEvent(c.ctx)
		if err != nil {
			return errors.Trace(err)
		}

		// Update the delay between the Canal and the Master before the handler hooks are called
		c.updateReplicationDelay(ev)

		// If log pos equals zero then the received event is a fake rotate event and
		// contains only a name of the next binlog file
		// See https://github.com/mysql/mysql-server/blob/8e797a5d6eb3a87f16498edcb7261a75897babae/sql/rpl_binlog_sender.h#L235
		// and https://github.com/mysql/mysql-server/blob/8cc757da3d87bf4a1f07dcfb2d3c96fed3806870/sql/rpl_binlog_sender.cc#L899
		if ev.Header.LogPos == 0 {
			switch e := ev.Event.(type) {
			case *replication.RotateEvent:
				fakeRotateLogName = string(e.NextLogName)
				log.Infof("received fake rotate event, next log name is %s", e.NextLogName)
			}

			continue
		}

		savePos = false
		force = false
		pos := c.master.Position()

		curPos := pos.Pos
		// next binlog pos
		pos.Pos = ev.Header.LogPos

		// new file name received in the fake rotate event
		if fakeRotateLogName != "" {
			pos.Name = fakeRotateLogName
		}

		// We only save position with RotateEvent and XIDEvent.
		// For RowsEvent, we can't save the position until meeting XIDEvent
		// which tells the whole transaction is over.
		// TODO: If we meet any DDL query, we must save too.
		switch e := ev.Event.(type) {
		case *replication.RotateEvent:
			pos.Name = string(e.NextLogName)
			pos.Pos = uint32(e.Position)
			log.Infof("rotate binlog to %s", pos)
			savePos = true
			force = true
			if err = c.eventHandler.OnRotate(e); err != nil {
				return errors.Trace(err)
			}
		case *replication.RowsEvent:
			// we only focus row based event
			err = c.handleRowsEvent(ev)
			if err != nil {
				e := errors.Cause(err)
				// if error is not ErrExcludedTable or ErrTableNotExist or ErrMissingTableMeta, stop canal
				if e != ErrExcludedTable &&
					e != schema.ErrTableNotExist &&
					e != schema.ErrMissingTableMeta {
					log.Errorf("handle rows event at (%s, %d) error %v", pos.Name, curPos, err)
					return errors.Trace(err)
				}
			}
			continue
		case *replication.XIDEvent:
			savePos = true
			// try to save the position later
			if err := c.eventHandler.OnXID(pos); err != nil {
				return errors.Trace(err)
			}
			if e.GSet != nil {
				c.master.UpdateGTIDSet(e.GSet)
			}
		case *replication.MariadbGTIDEvent:
			// try to save the GTID later
			gtid, err := mysql.ParseMariadbGTIDSet(e.GTID.String())
			if err != nil {
				return errors.Trace(err)
			}
			if err := c.eventHandler.OnGTID(gtid); err != nil {
				return errors.Trace(err)
			}
		case *replication.GTIDEvent:
			u, _ := uuid.FromBytes(e.SID)
			gtid, err := mysql.ParseMysqlGTIDSet(fmt.Sprintf("%s:%d", u.String(), e.GNO))
			if err != nil {
				return errors.Trace(err)
			}
			if err := c.eventHandler.OnGTID(gtid); err != nil {
				return errors.Trace(err)
			}
		case *replication.QueryEvent:
			stmts, _, err := c.parser.Parse(string(e.Query), "", "")
			if err != nil {
				log.Errorf("parse query(%s) err %v, will skip this event", e.Query, err)
				continue
			}
			for _, stmt := range stmts {
				nodes := parseStmt(stmt)
				for _, node := range nodes {
					if node.db == "" {
						node.db = string(e.Schema)
					}
					if err = c.updateTable(node.db, node.table); err != nil {
						return errors.Trace(err)
					}
				}
				if len(nodes) > 0 {
					savePos = true
					force = true
					// Now we only handle Table Changed DDL, maybe we will support more later.
					if err = c.eventHandler.OnDDL(pos, e); err != nil {
						return errors.Trace(err)
					}
				}
			}
			if savePos && e.GSet != nil {
				c.master.UpdateGTIDSet(e.GSet)
			}
		default:
			continue
		}

		if savePos {
			c.master.Update(pos)
			c.master.UpdateTimestamp(ev.Header.Timestamp)
			fakeRotateLogName = ""

			if err := c.eventHandler.OnPosSynced(pos, c.master.GTIDSet(), force); err != nil {
				return errors.Trace(err)
			}
		}
	}

	return nil
}
```
### 对cannal的使用主要依赖handler, 以下是handler的接口
```
type EventHandler interface {
	OnRotate(roateEvent *replication.RotateEvent) error
	// OnTableChanged is called when the table is created, altered, renamed or dropped.
	// You need to clear the associated data like cache with the table.
	// It will be called before OnDDL.
	OnTableChanged(schema string, table string) error
	OnBegin(pos mysql.Position, eh *replication.EventHeader) error
	OnDDL(ddlType int, db, table string, h *replication.EventHeader, queryEvent *replication.QueryEvent) error
	OnUnHandledQuery(h *replication.EventHeader, queryEvent *replication.QueryEvent, reason string) error
	OnRow(e *RowsEvent) error
	OnXID(h *replication.EventHeader, e *replication.XIDEvent, nextPos mysql.Position) error
	OnMariaDBGTID(e *replication.MariadbGTIDEvent) error
	OnGTID(e *replication.GTIDEvent) error
	// OnPosSynced Use your own way to sync position. When force is true, sync position immediately.
	OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error
	String() string
}
```

### Binlog中EventHeader的EventType
```
const (
	UNKNOWN_EVENT EventType = iota
	START_EVENT_V3
	QUERY_EVENT
	STOP_EVENT
	ROTATE_EVENT
	INTVAR_EVENT
	LOAD_EVENT
	SLAVE_EVENT
	CREATE_FILE_EVENT
	APPEND_BLOCK_EVENT
	EXEC_LOAD_EVENT
	DELETE_FILE_EVENT
	NEW_LOAD_EVENT
	RAND_EVENT
	USER_VAR_EVENT
	FORMAT_DESCRIPTION_EVENT
	XID_EVENT
	BEGIN_LOAD_QUERY_EVENT
	EXECUTE_LOAD_QUERY_EVENT
	TABLE_MAP_EVENT
	WRITE_ROWS_EVENTv0
	UPDATE_ROWS_EVENTv0
	DELETE_ROWS_EVENTv0
	WRITE_ROWS_EVENTv1
	UPDATE_ROWS_EVENTv1
	DELETE_ROWS_EVENTv1
	INCIDENT_EVENT
	HEARTBEAT_EVENT
	IGNORABLE_EVENT
	ROWS_QUERY_EVENT
	WRITE_ROWS_EVENTv2
	UPDATE_ROWS_EVENTv2
	DELETE_ROWS_EVENTv2
	GTID_EVENT
	ANONYMOUS_GTID_EVENT
	PREVIOUS_GTIDS_EVENT
	TRANSACTION_CONTEXT_EVENT
	VIEW_CHANGE_EVENT
	XA_PREPARE_LOG_EVENT
)
```

### 把Binlog的Event解析为RowEvent 和TableEvent
```
func (p *BinlogParser) parseEvent(h *EventHeader, data []byte, rawData []byte) (Event, error) {
	var e Event

	if h.EventType == FORMAT_DESCRIPTION_EVENT {
		p.format = &FormatDescriptionEvent{}
		e = p.format
	} else {
		if p.format != nil && p.format.ChecksumAlgorithm == BINLOG_CHECKSUM_ALG_CRC32 {
			err := p.verifyCrc32Checksum(rawData)
			if err != nil {
				return nil, err
			}
			data = data[0 : len(data)-BinlogChecksumLength]
		}

		if h.EventType == ROTATE_EVENT {
			e = &RotateEvent{}
		} else if !p.rawMode {
			switch h.EventType {
			case QUERY_EVENT:
				e = &QueryEvent{}
			case XID_EVENT:
				e = &XIDEvent{}
			case TABLE_MAP_EVENT:
				te := &TableMapEvent{}
				if p.format.EventTypeHeaderLengths[TABLE_MAP_EVENT-1] == 6 {
					te.tableIDSize = 4
				} else {
					te.tableIDSize = 6
				}
				e = te
			case WRITE_ROWS_EVENTv0,
				UPDATE_ROWS_EVENTv0,
				DELETE_ROWS_EVENTv0,
				WRITE_ROWS_EVENTv1,
				DELETE_ROWS_EVENTv1,
				UPDATE_ROWS_EVENTv1,
				WRITE_ROWS_EVENTv2,
				UPDATE_ROWS_EVENTv2,
				DELETE_ROWS_EVENTv2:
				e = p.newRowsEvent(h)
			case ROWS_QUERY_EVENT:
				e = &RowsQueryEvent{}
			case GTID_EVENT:
				e = &GTIDEvent{}
			case ANONYMOUS_GTID_EVENT:
				e = &GTIDEvent{}
			case BEGIN_LOAD_QUERY_EVENT:
				e = &BeginLoadQueryEvent{}
			case EXECUTE_LOAD_QUERY_EVENT:
				e = &ExecuteLoadQueryEvent{}
			case MARIADB_ANNOTATE_ROWS_EVENT:
				e = &MariadbAnnotateRowsEvent{}
			case MARIADB_BINLOG_CHECKPOINT_EVENT:
				e = &MariadbBinlogCheckPointEvent{}
			case MARIADB_GTID_LIST_EVENT:
				e = &MariadbGTIDListEvent{}
			case MARIADB_GTID_EVENT:
				ee := &MariadbGTIDEvent{}
				ee.GTID.ServerID = h.ServerID
				e = ee
			case PREVIOUS_GTIDS_EVENT:
				e = &PreviousGTIDsEvent{}
			default:
				e = &GenericEvent{}
			}
		} else {
			e = &GenericEvent{}
		}
	}

	if err := e.Decode(data); err != nil {
		return nil, &EventError{h, err.Error(), data}
	}

	if te, ok := e.(*TableMapEvent); ok {
		p.tables[te.TableID] = te
	}

	if re, ok := e.(*RowsEvent); ok {
		if (re.Flags & RowsEventStmtEndFlag) > 0 {
			// Refer https://github.com/alibaba/canal/blob/38cc81b7dab29b51371096fb6763ca3a8432ffee/dbsync/src/main/java/com/taobao/tddl/dbsync/binlog/event/RowsLogEvent.java#L176
			p.tables = make(map[uint64]*TableMapEvent)
		}
	}

	return e, nil
}
```
### Mysql type
// see mysql sql/log_event.h
