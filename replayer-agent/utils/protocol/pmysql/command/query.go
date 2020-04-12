package command

import (
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// QueryBody query body
type QueryBody struct {
	RawBody string
}

func (q *QueryBody) String() string {
	data, err := json.Marshal(q)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (q *QueryBody) Map() model.Map {
	r := make(model.Map)
	r["sql"] = q.RawBody
	return r
}

// DecodeQueryReq 解码query请求
// doc: https://dev.mysql.com/doc/internals/en/com-query.html
func DecodeQueryReq(src *parse.Source) (*QueryBody, error) {
	pkLen, _ := common.GetPacketHeader(src)
	b := src.Read1()
	if b != 0x03 {
		return nil, errors.New("packet isn't a query request")
	}
	return &QueryBody{RawBody: string(src.ReadN(pkLen - 1))}, nil
}

// ResultSet query的返回结果集
type ResultSet struct {
	Columns []Columndef
	DataSet [][]string
}

func (rs *ResultSet) String() string {
	data, err := json.Marshal(rs)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map 把self转成一个model.Map对象
func (rs *ResultSet) Map() model.Map {
	r := make(model.Map)
	list := make(model.List, 0, len(rs.Columns))
	for _, col := range rs.Columns {
		list = append(list, col.Map())
	}
	r["columns"] = list
	dataset := make(model.List, 0, len(rs.DataSet))
	for _, set := range rs.DataSet {
		dataset = append(dataset, set)
	}
	r["rows"] = list
	return r
}

// Columndef table的一个字段的定义
type Columndef struct {
	// 0x03 def
	Schema     string
	Table      string
	OrgTable   string
	ColName    string
	OrgColName string
	// next_length always 0x0c
	Charset    uint16 // 2bytes
	ColLength  int    // int<4>
	Type       byte
	ExtraBytes []byte
}

// Map 把self转成一个model.Map对象
func (colDef *Columndef) Map() model.Map {
	r := make(model.Map)
	r["schema"] = colDef.Schema
	r["table"] = colDef.Table
	r["origin_table"] = colDef.OrgTable
	r["column_name"] = colDef.ColName
	r["origin_column_name"] = colDef.OrgColName
	r["charset"] = colDef.Charset
	r["column_length"] = colDef.ColLength
	r["type"] = colDef.Type
	r["extra_bytes"] = colDef.ExtraBytes
	return r
}

// DecodeResultSet 解码resultset包
// doc: https://dev.mysql.com/doc/internals/en/com-query-response.html#packet-ProtocolText::Resultset
func DecodeResultSet(src *parse.Source) (*ResultSet, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	colNum, _, err := common.GetIntLenc(src.ReadN(pkLen))
	if nil != err {
		return nil, err
	}
	if colNum == 0 {
		return nil, errNoResultSetPacket
	}
	if src.Error() != nil {
		return nil, src.Error()
	}
	resultSet := &ResultSet{}
	for i := 0; i < colNum; i++ {
		field, err := DecodeColumnDef(src)
		if nil != err {
			return nil, err
		}
		resultSet.Columns = append(resultSet.Columns, field)
	}
	for src.Error() == nil && !common.IsEOFPacket(src) {
		dataSet, err := decodeDataset(src)
		if nil != err {
			return nil, err
		}
		resultSet.DataSet = append(resultSet.DataSet, dataSet)
	}
	if src.Error() != nil {
		return nil, src.Error()
	}
	common.ReadEOFPacket(src)
	// -- should end

	src.Peek1()
	// if there's more bytes, this isn't resultset packet
	if src.Error() == nil {
		return nil, errNoResultSetPacket
	} else {
		return resultSet, nil
	}
}

const nullStr = "NULL"

func decodeDataset(src *parse.Source) ([]string, error) {
	pkLen, _ := common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	var dataSet []string
	consumed := 0
	for consumed < pkLen {
		b := src.Peek1()
		if src.Error() != nil {
			return nil, src.Error()
		}
		if b == common.NULL {
			src.Read1()
			consumed++
			dataSet = append(dataSet, nullStr)
			continue
		}
		val := common.GetLenencString(src)
		if src.Error() != nil {
			return nil, src.Error()
		}
		dataSet = append(dataSet, val)
		consumed += common.GetLenencStringLength(val)
	}
	return dataSet, nil
}

// DecodeColumnDef 解码column定义
func DecodeColumnDef(src *parse.Source) (Columndef, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	// 03 def
	src.ReadN(4)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	schema := common.GetLenencString(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	table := common.GetLenencString(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	orgTable := common.GetLenencString(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	colName := common.GetLenencString(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	orgColName := common.GetLenencString(src)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	// 0x0c
	src.Read1()
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	charset, err := common.GetIntN(src.ReadN(2), 2)
	if nil != err {
		return Columndef{}, err
	}
	colLen, err := common.GetIntN(src.ReadN(4), 4)
	if nil != err {
		return Columndef{}, err
	}
	dataType := src.Read1()
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	// flags 2 + decimals 1 + filter 2(0x00 0x00)
	extraBytes := src.ReadN(5)
	if src.Error() != nil {
		return Columndef{}, src.Error()
	}
	return Columndef{
		Schema:     schema,
		Table:      table,
		OrgTable:   orgTable,
		ColName:    colName,
		OrgColName: orgColName,
		Charset:    uint16(charset),
		ColLength:  colLen,
		Type:       dataType,
		ExtraBytes: extraBytes,
	}, nil
}

var (
	errNoResultSetPacket = errors.New("not resultset packet")
)
