package replication

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pingcap/errors"
	uuid "github.com/satori/go.uuid"

	. "github.com/go-mysql-org/go-mysql/mysql"
)

const (
	EventHeaderSize            = 19
	SidLength                  = 16
	LogicalTimestampTypeCode   = 2
	PartLogicalTimestampLength = 8
	BinlogChecksumLength       = 4
	UndefinedServerVer         = 999999 // UNDEFINED_SERVER_VERSION
)

// These constants describe the type of status variables in q Query packet.
const (
	// QFlags2Code is Q_FLAGS2_CODE
	QFlags2Code = 0

	// QSQLModeCode is Q_SQL_MODE_CODE
	QSQLModeCode = 1

	// QCatalog is Q_CATALOG
	QCatalog = 2

	// QAutoIncrement is Q_AUTO_INCREMENT
	QAutoIncrement = 3

	// QCharsetCode is Q_CHARSET_CODE
	QCharsetCode = 4

	// QTimeZoneCode is Q_TIME_ZONE_CODE
	QTimeZoneCode = 5

	// QCatalogNZCode is Q_CATALOG_NZ_CODE
	QCatalogNZCode = 6

	// QLCTimeNamesCode is Q_LC_TIME_NAMES_CODE
	QLCTimeNamesCode = 7

	// QCharsetDatabaseCode is Q_CHARSET_DATABASE_CODE
	QCharsetDatabaseCode = 8

	// QTableMapForUpdateCode is Q_TABLE_MAP_FOR_UPDATE_CODE
	QTableMapForUpdateCode = 9

	// QMasterDataWrittenCode is Q_MASTER_DATA_WRITTEN_CODE
	QMasterDataWrittenCode = 10

	// QInvokers is Q_INVOKERS
	QInvokers = 11

	// QUpdatedDBNames is Q_UPDATED_DB_NAMES
	QUpdatedDBNames = 12

	// QMicroSeconds is Q_MICROSECONDS
	QMicroSeconds = 13
)

type BinlogEvent struct {
	// raw binlog data which contains all data, including binlog header and event body, and including crc32 checksum if exists
	RawData []byte

	Header *EventHeader
	Event  Event
}

func (e *BinlogEvent) Dump(w io.Writer) {
	e.Header.Dump(w)
	e.Event.Dump(w)
}

type Event interface {
	//Dump Event, format like python-mysql-replication
	Dump(w io.Writer)

	Decode(data []byte) error
}

type EventError struct {
	Header *EventHeader

	//Error message
	Err string

	//Event data
	Data []byte
}

func (e *EventError) Error() string {
	return fmt.Sprintf("Header %#v, Data %q, Err: %v", e.Header, e.Data, e.Err)
}

type EventHeader struct {
	Timestamp uint32
	EventType EventType
	ServerID  uint32
	EventSize uint32
	LogPos    uint32
	Flags     uint16
}

func (h *EventHeader) Decode(data []byte) error {
	if len(data) < EventHeaderSize {
		return errors.Errorf("header size too short %d, must 19", len(data))
	}

	pos := 0

	h.Timestamp = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	h.EventType = EventType(data[pos])
	pos++

	h.ServerID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	h.EventSize = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	h.LogPos = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	h.Flags = binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	if h.EventSize < uint32(EventHeaderSize) {
		return errors.Errorf("invalid event size %d, must >= 19", h.EventSize)
	}

	return nil
}

func (h *EventHeader) Dump(w io.Writer) {
	fmt.Fprintf(w, "=== %s ===\n", h.EventType)
	fmt.Fprintf(w, "Date: %s\n", time.Unix(int64(h.Timestamp), 0).Format(TimeFormat))
	fmt.Fprintf(w, "Log position: %d\n", h.LogPos)
	fmt.Fprintf(w, "Event size: %d\n", h.EventSize)
}

var (
	checksumVersionSplitMysql   []int = []int{5, 6, 1}
	checksumVersionProductMysql int   = (checksumVersionSplitMysql[0]*256+checksumVersionSplitMysql[1])*256 + checksumVersionSplitMysql[2]

	checksumVersionSplitMariaDB   []int = []int{5, 3, 0}
	checksumVersionProductMariaDB int   = (checksumVersionSplitMariaDB[0]*256+checksumVersionSplitMariaDB[1])*256 + checksumVersionSplitMariaDB[2]
)

// server version format X.Y.Zabc, a is not . or number
func splitServerVersion(server string) []int {
	seps := strings.Split(server, ".")
	if len(seps) < 3 {
		return []int{0, 0, 0}
	}

	x, _ := strconv.Atoi(seps[0])
	y, _ := strconv.Atoi(seps[1])

	index := 0
	for i, c := range seps[2] {
		if !unicode.IsNumber(c) {
			index = i
			break
		}
	}

	z, _ := strconv.Atoi(seps[2][0:index])

	return []int{x, y, z}
}

func calcVersionProduct(server string) int {
	versionSplit := splitServerVersion(server)

	return ((versionSplit[0]*256+versionSplit[1])*256 + versionSplit[2])
}

type FormatDescriptionEvent struct {
	Version uint16
	//len = 50
	ServerVersion          []byte
	CreateTimestamp        uint32
	EventHeaderLength      uint8
	EventTypeHeaderLengths []byte

	// 0 is off, 1 is for CRC32, 255 is undefined
	ChecksumAlgorithm byte
}

func (e *FormatDescriptionEvent) Decode(data []byte) error {
	pos := 0
	e.Version = binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	e.ServerVersion = make([]byte, 50)
	copy(e.ServerVersion, data[pos:])
	pos += 50

	e.CreateTimestamp = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.EventHeaderLength = data[pos]
	pos++

	if e.EventHeaderLength != byte(EventHeaderSize) {
		return errors.Errorf("invalid event header length %d, must 19", e.EventHeaderLength)
	}

	server := string(e.ServerVersion)
	checksumProduct := checksumVersionProductMysql
	if strings.Contains(strings.ToLower(server), "mariadb") {
		checksumProduct = checksumVersionProductMariaDB
	}

	if calcVersionProduct(string(e.ServerVersion)) >= checksumProduct {
		// here, the last 5 bytes is 1 byte check sum alg type and 4 byte checksum if exists
		e.ChecksumAlgorithm = data[len(data)-5]
		e.EventTypeHeaderLengths = data[pos : len(data)-5]
	} else {
		e.ChecksumAlgorithm = BINLOG_CHECKSUM_ALG_UNDEF
		e.EventTypeHeaderLengths = data[pos:]
	}

	return nil
}

func (e *FormatDescriptionEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Version: %d\n", e.Version)
	fmt.Fprintf(w, "Server version: %s\n", e.ServerVersion)
	//fmt.Fprintf(w, "Create date: %s\n", time.Unix(int64(e.CreateTimestamp), 0).Format(TimeFormat))
	fmt.Fprintf(w, "Checksum algorithm: %d\n", e.ChecksumAlgorithm)
	//fmt.Fprintf(w, "Event header lengths: \n%s", hex.Dump(e.EventTypeHeaderLengths))
	fmt.Fprintln(w)
}

type RotateEvent struct {
	Position    uint64
	NextLogName []byte
}

func (e *RotateEvent) Decode(data []byte) error {
	e.Position = binary.LittleEndian.Uint64(data[0:])
	e.NextLogName = data[8:]

	return nil
}

func (e *RotateEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Position: %d\n", e.Position)
	fmt.Fprintf(w, "Next log name: %s\n", e.NextLogName)
	fmt.Fprintln(w)
}

type PreviousGTIDsEvent struct {
	GTIDSets string
}

func (e *PreviousGTIDsEvent) Decode(data []byte) error {
	var previousGTIDSets []string
	pos := 0
	uuidCount := binary.LittleEndian.Uint16(data[pos : pos+8])
	pos += 8

	for i := uint16(0); i < uuidCount; i++ {
		uuid := e.decodeUuid(data[pos : pos+16])
		pos += 16
		sliceCount := binary.LittleEndian.Uint16(data[pos : pos+8])
		pos += 8
		var intervals []string
		for i := uint16(0); i < sliceCount; i++ {
			start := e.decodeInterval(data[pos : pos+8])
			pos += 8
			stop := e.decodeInterval(data[pos : pos+8])
			pos += 8
			interval := ""
			if stop == start+1 {
				interval = fmt.Sprintf("%d", start)
			} else {
				interval = fmt.Sprintf("%d-%d", start, stop-1)
			}
			intervals = append(intervals, interval)
		}
		previousGTIDSets = append(previousGTIDSets, fmt.Sprintf("%s:%s", uuid, strings.Join(intervals, ":")))
	}
	e.GTIDSets = fmt.Sprintf("%s", strings.Join(previousGTIDSets, ","))
	return nil
}

func (e *PreviousGTIDsEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Previous GTID Event: %s\n", e.GTIDSets)
	fmt.Fprintln(w)
}

func (e *PreviousGTIDsEvent) decodeUuid(data []byte) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(data[0:4]), hex.EncodeToString(data[4:6]),
		hex.EncodeToString(data[6:8]), hex.EncodeToString(data[8:10]), hex.EncodeToString(data[10:]))
}

func (e *PreviousGTIDsEvent) decodeInterval(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

type XIDEvent struct {
	XID uint64

	// in fact XIDEvent dosen't have the GTIDSet information, just for beneficial to use
	GSet GTIDSet
}

func (e *XIDEvent) Decode(data []byte) error {
	e.XID = binary.LittleEndian.Uint64(data)
	return nil
}

func (e *XIDEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "XID: %d\n", e.XID)
	if e.GSet != nil {
		fmt.Fprintf(w, "GTIDSet: %s\n", e.GSet.String())
	}
	fmt.Fprintln(w)
}

//Charset 字符集（客户端，连接，服务器）
type Charset struct {
	// @@session.character_set_client
	Client uint16
	// @@session.collation_connection
	Conn uint16
	// @@session.collation_server
	Server uint16
}

type QueryEvent struct {
	SlaveProxyID  uint32
	ExecutionTime uint32
	ErrorCode     uint16
	StatusVars    []byte
	Schema        []byte
	Query         []byte

	SqlMode  *uint64
	Charset  *Charset
	Timezone *string
	// @@session.collation_database
	DatabaseCharset *uint16
	// in fact QueryEvent dosen't have the GTIDSet information, just for beneficial to use
	GSet GTIDSet
}

func (e *QueryEvent) Decode(data []byte) error {
	pos := 0

	e.SlaveProxyID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.ExecutionTime = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	schemaLength := data[pos]
	pos++

	e.ErrorCode = binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	statusVarsLength := binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	e.StatusVars = data[pos : pos+int(statusVarsLength)]
	pos += int(statusVarsLength)

	e.Schema = data[pos : pos+int(schemaLength)]
	pos += int(schemaLength)

	//skip 0x00
	pos++

	e.Query = data[pos:]
	e.ParseStatusVars()
	return nil
}

func (e *QueryEvent) ParseStatusVars() error {
	vars := e.StatusVars
	for pos := 0; pos < len(vars); {
		code := vars[pos]
		pos++

		// All codes are optional, but if present they must occur in numerically
		// increasing order (except for 6 which occurs in the place of 2) to allow
		// for backward compatibility.
		switch code {
		case QFlags2Code, QAutoIncrement:
			pos += 4
		case QSQLModeCode:
			e.SqlMode = new(uint64)
			*e.SqlMode = binary.LittleEndian.Uint64(vars[pos : pos+8])
			pos += 8
		case QCatalog: // Used in MySQL 5.0.0 - 5.0.3
			if pos+1 > len(vars) {
				return fmt.Errorf("Q_CATALOG status var overflows buffer (%v + 1 > %v)", pos, len(vars))
			}
			pos += 1 + int(vars[pos]) + 1
		case QCatalogNZCode: // Used in MySQL > 5.0.3 to replace QCatalog
			if pos+1 > len(vars) {
				return fmt.Errorf("Q_CATALOG_NZ_CODE status var overflows buffer (%v + 1 > %v)", pos, len(vars))
			}
			pos += 1 + int(vars[pos])
		case QCharsetCode:
			if pos+6 > len(vars) {
				return fmt.Errorf("Q_CHARSET_CODE status var overflows buffer (%v + 6 > %v)", pos, len(vars))
			}
			e.Charset = &Charset{
				Client: binary.LittleEndian.Uint16(vars[pos : pos+2]),
				Conn:   binary.LittleEndian.Uint16(vars[pos+2 : pos+4]),
				Server: binary.LittleEndian.Uint16(vars[pos+4 : pos+6]),
			}
			pos += 6
		case QTimeZoneCode:
			if pos+1 > len(vars) {
				return fmt.Errorf("QTimeZoneCode status var overflows buffer (%v + 1 > %v)", pos, len(vars))
			}
			e.Timezone = new(string)
			*e.Timezone = string(vars[pos+1 : pos+1+int(vars[pos])])
			pos += 1 + int(vars[pos])
		case QLCTimeNamesCode:
			pos += 2
		case QCharsetDatabaseCode:
			e.DatabaseCharset = new(uint16)
			*e.DatabaseCharset = binary.LittleEndian.Uint16(vars[pos : pos+2])
			pos += 2
		case QTableMapForUpdateCode:
			pos += 8
		case QMasterDataWrittenCode:
			pos += 4
		case QInvokers:
			if pos+1 > len(vars) {
				return fmt.Errorf("QInvokers status var overflows buffer (%v + 1 > %v)", pos, len(vars))
			}
			pos += 1 + int(vars[pos])
			if pos+1 > len(vars) {
				return fmt.Errorf("QInvokers status var overflows buffer (%v + 1 > %v)", pos, len(vars))
			}
			pos += 1 + int(vars[pos])
		case QUpdatedDBNames:
			n := int(vars[pos])
			pos++
			for i := 0; i < n; i++ {
				idx := bytes.IndexByte(vars[pos:], byte(0))
				if idx > -1 {
					pos += idx + 1
				} else {
					return errors.New("QUpdatedDBNames: Not enough data")
				}
			}
		case QMicroSeconds:
			pos += 3
		default:
			// If we see something higher than what we're interested in, we can stop.
			return nil
		}
	}
	return nil
}

func (e *QueryEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Slave proxy ID: %d\n", e.SlaveProxyID)
	fmt.Fprintf(w, "Execution time: %d\n", e.ExecutionTime)
	fmt.Fprintf(w, "Error code: %d\n", e.ErrorCode)
	//fmt.Fprintf(w, "Status vars: \n%s", hex.Dump(e.StatusVars))
	fmt.Fprintf(w, "Schema: %s\n", e.Schema)
	fmt.Fprintf(w, "Query: %s\n", e.Query)
	if e.GSet != nil {
		fmt.Fprintf(w, "GTIDSet: %s\n", e.GSet.String())
	}
	fmt.Fprintln(w)
}

type GTIDEvent struct {
	CommitFlag     uint8
	SID            []byte
	GNO            int64
	LastCommitted  int64
	SequenceNumber int64

	// ImmediateCommitTimestamp/OriginalCommitTimestamp are introduced in MySQL-8.0.1, see:
	// https://mysqlhighavailability.com/replication-features-in-mysql-8-0-1/
	ImmediateCommitTimestamp uint64
	OriginalCommitTimestamp  uint64

	// Total transaction length (including this GTIDEvent), introduced in MySQL-8.0.2, see:
	// https://mysqlhighavailability.com/taking-advantage-of-new-transaction-length-metadata/
	TransactionLength uint64

	// ImmediateServerVersion/OriginalServerVersion are introduced in MySQL-8.0.14, see
	// https://dev.mysql.com/doc/refman/8.0/en/replication-compatibility.html
	ImmediateServerVersion uint32
	OriginalServerVersion  uint32
}

func (e *GTIDEvent) Decode(data []byte) error {
	pos := 0
	e.CommitFlag = data[pos]
	pos++
	e.SID = data[pos : pos+SidLength]
	pos += SidLength
	e.GNO = int64(binary.LittleEndian.Uint64(data[pos:]))
	pos += 8

	if len(data) >= 42 {
		if data[pos] == LogicalTimestampTypeCode {
			pos++
			e.LastCommitted = int64(binary.LittleEndian.Uint64(data[pos:]))
			pos += PartLogicalTimestampLength
			e.SequenceNumber = int64(binary.LittleEndian.Uint64(data[pos:]))
			pos += 8

			// IMMEDIATE_COMMIT_TIMESTAMP_LENGTH = 7
			if len(data)-pos < 7 {
				return nil
			}
			e.ImmediateCommitTimestamp = FixedLengthInt(data[pos : pos+7])
			pos += 7
			if (e.ImmediateCommitTimestamp & (uint64(1) << 55)) != 0 {
				// If the most significant bit set, another 7 byte follows representing OriginalCommitTimestamp
				e.ImmediateCommitTimestamp &= ^(uint64(1) << 55)
				e.OriginalCommitTimestamp = FixedLengthInt(data[pos : pos+7])
				pos += 7
			} else {
				// Otherwise OriginalCommitTimestamp == ImmediateCommitTimestamp
				e.OriginalCommitTimestamp = e.ImmediateCommitTimestamp
			}

			// TRANSACTION_LENGTH_MIN_LENGTH = 1
			if len(data)-pos < 1 {
				return nil
			}
			var n int
			e.TransactionLength, _, n = LengthEncodedInt(data[pos:])
			pos += n

			// IMMEDIATE_SERVER_VERSION_LENGTH = 4
			e.ImmediateServerVersion = UndefinedServerVer
			e.OriginalServerVersion = UndefinedServerVer
			if len(data)-pos < 4 {
				return nil
			}
			e.ImmediateServerVersion = binary.LittleEndian.Uint32(data[pos:])
			pos += 4
			if (e.ImmediateServerVersion & (uint32(1) << 31)) != 0 {
				// If the most significant bit set, another 4 byte follows representing OriginalServerVersion
				e.ImmediateServerVersion &= ^(uint32(1) << 31)
				e.OriginalServerVersion = binary.LittleEndian.Uint32(data[pos:])
				pos += 4
			} else {
				// Otherwise OriginalServerVersion == ImmediateServerVersion
				e.OriginalServerVersion = e.ImmediateServerVersion
			}
		}
	}
	return nil
}

func (e *GTIDEvent) Dump(w io.Writer) {
	fmtTime := func(t time.Time) string {
		if t.IsZero() {
			return "<n/a>"
		}
		return t.Format(time.RFC3339Nano)
	}

	fmt.Fprintf(w, "Commit flag: %d\n", e.CommitFlag)
	u, _ := uuid.FromBytes(e.SID)
	fmt.Fprintf(w, "GTID_NEXT: %s:%d\n", u.String(), e.GNO)
	fmt.Fprintf(w, "LAST_COMMITTED: %d\n", e.LastCommitted)
	fmt.Fprintf(w, "SEQUENCE_NUMBER: %d\n", e.SequenceNumber)
	fmt.Fprintf(w, "Immediate commmit timestamp: %d (%s)\n", e.ImmediateCommitTimestamp, fmtTime(e.ImmediateCommitTime()))
	fmt.Fprintf(w, "Orignal commmit timestamp: %d (%s)\n", e.OriginalCommitTimestamp, fmtTime(e.OriginalCommitTime()))
	fmt.Fprintf(w, "Transaction length: %d\n", e.TransactionLength)
	fmt.Fprintf(w, "Immediate server version: %d\n", e.ImmediateServerVersion)
	fmt.Fprintf(w, "Orignal server version: %d\n", e.OriginalServerVersion)
	fmt.Fprintln(w)
}

// ImmediateCommitTime returns the commit time of this trx on the immediate server
// or zero time if not available.
func (e *GTIDEvent) ImmediateCommitTime() time.Time {
	return microSecTimestampToTime(e.ImmediateCommitTimestamp)
}

// OriginalCommitTime returns the commit time of this trx on the original server
// or zero time if not available.
func (e *GTIDEvent) OriginalCommitTime() time.Time {
	return microSecTimestampToTime(e.OriginalCommitTimestamp)
}

type BeginLoadQueryEvent struct {
	FileID    uint32
	BlockData []byte
}

func (e *BeginLoadQueryEvent) Decode(data []byte) error {
	pos := 0

	e.FileID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.BlockData = data[pos:]

	return nil
}

func (e *BeginLoadQueryEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "File ID: %d\n", e.FileID)
	fmt.Fprintf(w, "Block data: %s\n", e.BlockData)
	fmt.Fprintln(w)
}

type ExecuteLoadQueryEvent struct {
	SlaveProxyID     uint32
	ExecutionTime    uint32
	SchemaLength     uint8
	ErrorCode        uint16
	StatusVars       uint16
	FileID           uint32
	StartPos         uint32
	EndPos           uint32
	DupHandlingFlags uint8
}

func (e *ExecuteLoadQueryEvent) Decode(data []byte) error {
	pos := 0

	e.SlaveProxyID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.ExecutionTime = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.SchemaLength = data[pos]
	pos++

	e.ErrorCode = binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	e.StatusVars = binary.LittleEndian.Uint16(data[pos:])
	pos += 2

	e.FileID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.StartPos = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.EndPos = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	e.DupHandlingFlags = data[pos]

	return nil
}

func (e *ExecuteLoadQueryEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Slave proxy ID: %d\n", e.SlaveProxyID)
	fmt.Fprintf(w, "Execution time: %d\n", e.ExecutionTime)
	fmt.Fprintf(w, "Schame length: %d\n", e.SchemaLength)
	fmt.Fprintf(w, "Error code: %d\n", e.ErrorCode)
	fmt.Fprintf(w, "Status vars length: %d\n", e.StatusVars)
	fmt.Fprintf(w, "File ID: %d\n", e.FileID)
	fmt.Fprintf(w, "Start pos: %d\n", e.StartPos)
	fmt.Fprintf(w, "End pos: %d\n", e.EndPos)
	fmt.Fprintf(w, "Dup handling flags: %d\n", e.DupHandlingFlags)
	fmt.Fprintln(w)
}

// case MARIADB_ANNOTATE_ROWS_EVENT:
// 	return "MariadbAnnotateRowsEvent"

type MariadbAnnotateRowsEvent struct {
	Query []byte
}

func (e *MariadbAnnotateRowsEvent) Decode(data []byte) error {
	e.Query = data
	return nil
}

func (e *MariadbAnnotateRowsEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Query: %s\n", e.Query)
	fmt.Fprintln(w)
}

type MariadbBinlogCheckPointEvent struct {
	Info []byte
}

func (e *MariadbBinlogCheckPointEvent) Decode(data []byte) error {
	e.Info = data
	return nil
}

func (e *MariadbBinlogCheckPointEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Info: %s\n", e.Info)
	fmt.Fprintln(w)
}

type MariadbGTIDEvent struct {
	GTID     MariadbGTID
	Flags    byte
	CommitID uint64
}

func (e *MariadbGTIDEvent) IsDDL() bool {
	return (e.Flags & BINLOG_MARIADB_FL_DDL) != 0
}

func (e *MariadbGTIDEvent) IsStandalone() bool {
	return (e.Flags & BINLOG_MARIADB_FL_STANDALONE) != 0
}

func (e *MariadbGTIDEvent) IsGroupCommit() bool {
	return (e.Flags & BINLOG_MARIADB_FL_GROUP_COMMIT_ID) != 0
}

func (e *MariadbGTIDEvent) Decode(data []byte) error {
	pos := 0
	e.GTID.SequenceNumber = binary.LittleEndian.Uint64(data)
	pos += 8
	e.GTID.DomainID = binary.LittleEndian.Uint32(data[pos:])
	pos += 4
	e.Flags = data[pos]
	pos += 1

	if (e.Flags & BINLOG_MARIADB_FL_GROUP_COMMIT_ID) > 0 {
		e.CommitID = binary.LittleEndian.Uint64(data[pos:])
	}

	return nil
}

func (e *MariadbGTIDEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "GTID: %v\n", e.GTID)
	fmt.Fprintf(w, "Flags: %v\n", e.Flags)
	fmt.Fprintf(w, "CommitID: %v\n", e.CommitID)
	fmt.Fprintln(w)
}

type MariadbGTIDListEvent struct {
	GTIDs []MariadbGTID
}

func (e *MariadbGTIDListEvent) Decode(data []byte) error {
	pos := 0
	v := binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	count := v & uint32((1<<28)-1)

	e.GTIDs = make([]MariadbGTID, count)

	for i := uint32(0); i < count; i++ {
		e.GTIDs[i].DomainID = binary.LittleEndian.Uint32(data[pos:])
		pos += 4
		e.GTIDs[i].ServerID = binary.LittleEndian.Uint32(data[pos:])
		pos += 4
		e.GTIDs[i].SequenceNumber = binary.LittleEndian.Uint64(data[pos:])
		pos += 8
	}

	return nil
}

func (e *MariadbGTIDListEvent) Dump(w io.Writer) {
	fmt.Fprintf(w, "Lists: %v\n", e.GTIDs)
	fmt.Fprintln(w)
}
