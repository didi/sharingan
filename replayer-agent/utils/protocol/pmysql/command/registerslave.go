package command

import (
	"errors"

	"github.com/didi/sharingan/replayer-agent/utils/protocol/pmysql/common"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

// RegisterSlaveBody ...
type RegisterSlaveBody struct {
	ServerID         int
	SlaveHostNameLen int
	SlaveHostName    string
	SlaveUserLen     int
	SlaveUser        string
	SlavePasswordLen int
	SlavePassword    string
	SlaveMysqlPort   int
	ReplicationRank  int
	MasterID         int
}

func (p *RegisterSlaveBody) String() string {
	data, err := json.Marshal(p)
	if nil != err {
		return err.Error()
	}
	return string(data)
}

// Map ...
func (p *RegisterSlaveBody) Map() model.Map {
	r := make(model.Map)
	r["server_id"] = p.ServerID
	r["slave_hostname_len"] = p.SlaveHostNameLen
	r["slave_hostname"] = p.SlaveHostName
	r["slave_user_len"] = p.SlaveUserLen
	r["slave_user"] = p.SlaveUser
	r["slave_password_len"] = p.SlavePasswordLen
	r["slave_password"] = p.SlavePassword
	r["slave_mysql_port"] = p.SlaveMysqlPort
	r["replication_rank"] = p.ReplicationRank
	r["master_id"] = p.MasterID
	return r
}

// DecodeRegisterSlaveReq
// doc: https://dev.mysql.com/doc/internals/en/com-register-slave.html
func DecodeRegisterSlaveReq(src *parse.Source) (*RegisterSlaveBody, error) {
	common.GetPacketHeader(src)
	if src.Error() != nil {
		return nil, src.Error()
	}
	b := src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	if b != 0x15 {
		return nil, errors.New("packet isn't register slave query")
	}
	resp := new(RegisterSlaveBody)
	var err error
	resp.ServerID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.SlaveHostNameLen, err = common.GetIntN(src.ReadN(1), 1)
	if err != nil {
		return nil, err
	}
	resp.SlaveHostName = common.GetLenencString(src)
	resp.SlaveUserLen, err = common.GetIntN(src.ReadN(1), 1)
	if err != nil {
		return nil, err
	}
	resp.SlaveUser = common.GetLenencString(src)
	resp.SlavePasswordLen, err = common.GetIntN(src.ReadN(1), 1)
	if err != nil {
		return nil, err
	}
	resp.SlavePassword = common.GetLenencString(src)
	resp.SlaveMysqlPort, err = common.GetIntN(src.ReadN(2), 2)
	if err != nil {
		return nil, err
	}
	resp.ReplicationRank, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	resp.MasterID, err = common.GetIntN(src.ReadN(4), 4)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
