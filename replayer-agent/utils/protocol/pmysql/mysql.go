package pmysql

import (
	"bytes"
	"errors"

	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/command"
	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/handshake"
	"github.com/didichuxing/sharingan/replayer-agent/utils/protocol/pmysql/prepared"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

type packetType byte

// doc1: https://www.cnblogs.com/davygeek/p/5647175.html
// doc2: https://dev.mysql.com/doc/internals/en/client-server-protocol.html
func (t packetType) String() string {
	switch t {
	case _queryPacket:
		return "query请求"
	case _okPacket:
		return "OK返回"
	case _eofPacket:
		return "EOF包"
	case _errPacket:
		return "Err包"
	case _resultSetPacket:
		return "查询结果集"
	case _serverGreetingPacket:
		return "ServerGreeting"
	case _clientAuthPacket:
		return "clientAuth"
	case _prepareQuery:
		return "prepare query"
	case _prepareResponse:
		return "prepare response"
	case _stmtExecute:
		return "stmt execute"
	case _stmtExecuteResponse:
		return "stmt execute response"
	case _stmtClose:
		return "stmt close"
	case _stmtReset:
		return "stmt reset"
	case _sleepPacket:
		return "com sleep请求"
	case _quitPacket:
		return "com quit请求"
	case _initDBPacket:
		return "com init db请求"
	case _fieldListPacket:
		return "com field list请求"
	case _createDBPacket:
		return "com create db请求"
	case _dropDBPacket:
		return "com drop db请求"
	case _refreshPacket:
		return "com refresh请求"
	case _shutdownPacket:
		return "com shutdown请求"
	case _statisticsPacket:
		return "com statistics请求"
	case _processPacket:
		return "com process info请求"
	case _connectPacket:
		return "com connect请求"
	case _connectOutPacket:
		return "com connect out请求"
	case _processKillPacket:
		return "com process kill请求"
	case _debugPacket:
		return "com debug请求"
	case _pingPacket:
		return "com ping请求"
	case _timePacket:
		return "com time请求"
	case _delayedInsertPacket:
		return "com delayed insert请求"
	case _changeUserPacket:
		return "com change user请求"
	case _resetConnectionPacket:
		return "com reset connection请求"
	case _daemonPacket:
		return "com daemon请求"
	case _sendLongDataPacket:
		return "com send long data请求"
	case _setOptionPacket:
		return "com set option请求"
	case _fetchPacket:
		return "com fetch请求"
	case _binlogDumpPacket:
		return "com binlog dump请求"
	case _binlogDumpGtidPacket:
		return "com binlog dump GTID请求"
	case _tableDumpPacket:
		return "com table dump请求"
	case _registerSlavePacket:
		return "com register slave请求"
	default:
		return "未知包类型"
	}
}

const (
	_unknowPacket packetType = iota
	_queryPacket
	_okPacket
	_eofPacket
	_errPacket
	_resultSetPacket
	_serverGreetingPacket
	_clientAuthPacket
	_prepareQuery
	_prepareResponse
	_stmtExecute
	_stmtExecuteResponse
	_stmtClose
	_stmtReset
	_sleepPacket
	_quitPacket
	_initDBPacket
	_fieldListPacket
	_createDBPacket
	_dropDBPacket
	_refreshPacket
	_shutdownPacket
	_statisticsPacket
	_processPacket
	_connectPacket
	_connectOutPacket
	_processKillPacket
	_debugPacket
	_pingPacket
	_timePacket
	_delayedInsertPacket
	_changeUserPacket
	_resetConnectionPacket
	_daemonPacket
	_sendLongDataPacket
	_setOptionPacket
	_fetchPacket
	_binlogDumpPacket
	_binlogDumpGtidPacket
	_tableDumpPacket
	_registerSlavePacket
)

// DecodePacket 尝试用mysql协议解包，data是完整的抓包数据，包含TCP/IP头
func DecodePacket(data []byte) model.Map {
	payload, err := GetTCPPayload(data)
	if nil != err {
		return nil
	}
	return DecodePacketWithoutHeader(payload)
}

// DecodePacketWithoutHeader 尝试用Mysql协议解析，data是协议中应用层部分，不包含IP头等信息
func DecodePacketWithoutHeader(data []byte) model.Map {

	// 尝试顺序按可能性大小来排
	reader := bytes.NewReader(data)
	src, err := parse.NewSource(reader, 30)
	if nil != err {
		return nil
	}

	for _, protocol := range targetProtocol {
		src.StoreSavepoint()
		result, err := safeExecute(src, protocol)
		if nil == err && result != nil {
			return result
		}
		src.RollbackToSavepoint()
	}
	return nil
}

func safeExecute(src *parse.Source, fn protocolResolver) (mp model.Map, err error) {
	defer func() {
		if v := recover(); v != nil {
			mp = nil
			err = errors.New("unsupported mysql protocol")
		}
	}()
	mp, err = fn(src)
	return
}

type protocolResolver func(*parse.Source) (model.Map, error)

var targetProtocol = []protocolResolver{
	decodeQueryReq,
	decodeResultSet,
	decodeOKPacket,
	decodeEOFPacket,
	decodePrepareQuery,
	decodePrepareResponse,
	decodeStmtClose,
	decodeStmtReset,
	decodeStmtExecute,
	decodeStmtExecuteResponse,
	decodeErrPacket,
	decodeSleepReq,
	decodeQuitReq,
	decodeInitDBReq,
	decodeCreateDBReq,
	decodeDropDBReq,
	decodeRefreshReq,
	decodeShutdownReq,
	decodeStatisticsReq,
	decodeProcessReq,
	decodeConnectReq,
	decodeConnectOutReq,
	decodeDebugReq,
	decodePingReq,
	decodeTimeReq,
	decodeDelayedInsertReq,
	decodeChangeUserReq,
	decodeResetConnectionReq,
	decodeProcessKillReq,
	decodeDaemonReq,
	decodeSendLongDataReq,
	decodeBinlogDumpReq,
	decodeBinlogDumpGtidReq,
	decodeTableDumpReq,
	decodeSetOptionReq,
	decodeFetchReq,
	decodeFieldListReq,
	decodeServerGreeting,
	decodeClientLogin,
	decodeRegisterSlaveReq,
}

func decodeQueryReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeQueryReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _queryPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeResultSet(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeResultSet(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _resultSetPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeOKPacket(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeOKPacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _okPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeEOFPacket(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeEOFPacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _eofPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeErrPacket(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeErrPacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _errPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeSleepReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeSleepReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _sleepPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeQuitReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeQuitReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _quitPacket
	r["data"] = body.Map()
	return r, nil
}

func decodePingReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodePingReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _pingPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeTimeReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeTimeReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _timePacket
	r["data"] = body.Map()
	return r, nil
}

func decodeDaemonReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeDaemonReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _daemonPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeDebugReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeDebugReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _debugPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeFetchReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeFetchReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _fetchPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeDelayedInsertReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeDelayedInsertReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _delayedInsertPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeChangeUserReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeChangeUserReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _changeUserPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeBinlogDumpReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeBinlogDumpReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _binlogDumpPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeBinlogDumpGtidReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeBinlogDumpGtidReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _binlogDumpGtidPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeTableDumpReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeTableDumpReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _tableDumpPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeResetConnectionReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeResetConnectionReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _resetConnectionPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeInitDBReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeInitDBReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _initDBPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeCreateDBReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeCreateDBReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _createDBPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeRefreshReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeRefreshReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _refreshPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeSetOptionReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeSetOptionReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _setOptionPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeShutdownReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeShutdownReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _shutdownPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeStatisticsReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeStatisticsReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _statisticsPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeProcessReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeProcessReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _processPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeProcessKillReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeProcessKillReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _processKillPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeSendLongDataReq(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodeSendLongDataQuery(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _sendLongDataPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeConnectReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeConnectReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _connectPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeConnectOutReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeConnectOutReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _connectOutPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeDropDBReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeDropDBReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _dropDBPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeFieldListReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeFieldListReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _fieldListPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeServerGreeting(src *parse.Source) (model.Map, error) {
	body, err := handshake.DecodeServerGreeting(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _serverGreetingPacket
	r["data"] = body.Map()
	return r, nil
}

func decodeClientLogin(src *parse.Source) (model.Map, error) {
	body, err := handshake.DecodeClientLoginPacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _clientAuthPacket
	r["data"] = body.Map()
	return r, nil
}

func decodePrepareQuery(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodePreparedQuery(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _prepareQuery
	r["data"] = body.Map()
	return r, nil
}

func decodePrepareResponse(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodePrepareResponse(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _prepareResponse
	r["data"] = body.Map()
	return r, nil
}

func decodeStmtClose(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodeStmtClose(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _stmtClose
	r["data"] = body.Map()
	return r, nil
}

func decodeStmtReset(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodeStmtReset(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _stmtReset
	r["data"] = body.Map()
	return r, nil
}

func decodeStmtExecute(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodeStmtExecutePacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _stmtExecute
	r["data"] = body.Map()
	return r, nil
}

func decodeStmtExecuteResponse(src *parse.Source) (model.Map, error) {
	body, err := prepared.DecodeStmtExecuteResonsePacket(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _stmtExecuteResponse
	r["data"] = body.Map()
	return r, nil
}

func decodeRegisterSlaveReq(src *parse.Source) (model.Map, error) {
	body, err := command.DecodeRegisterSlaveReq(src)
	if nil != err {
		return nil, err
	}
	r := make(model.Map)
	r["protocol_type"] = _registerSlavePacket
	r["data"] = body.Map()
	return r, nil
}
