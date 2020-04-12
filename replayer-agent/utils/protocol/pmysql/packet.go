package pmysql

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// GetTCPPayload unpack TCP/IP packet
// return the payload without protocol layer
func GetTCPPayload(data []byte) ([]byte, error) {
	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
	if err := packet.ErrorLayer(); err != nil {
		return nil, err.Error()
	}
	appLayer := packet.ApplicationLayer()
	return appLayer.LayerContents(), nil
}
