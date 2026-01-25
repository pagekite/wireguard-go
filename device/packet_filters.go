package device

// FIXME: Document the why and how

import ()

type PacketDirection int

type PacketFilter interface {
	FilterFromTun([]byte, int64) []byte
	FilterToTun([]byte, *Peer, int64) []byte
}

type DefaultPacketFilter struct {
}

func MakeDefaultPacketFilter() PacketFilter {
	return &DefaultPacketFilter{}
}

func (pf *DefaultPacketFilter) FilterFromTun(packet []byte, nowms int64) []byte {
	return packet
}

func (pf *DefaultPacketFilter) FilterToTun(packet []byte, peer *Peer, nowms int64) []byte {
	return packet
}

func (device *Device) SendPacketsToPeer(peer *Peer, packets [][]byte) {
	elemsForPeer := device.GetOutboundElementsContainer()
	for _, packet := range packets {
		elem := device.NewOutboundElement()
		elem.packet = packet
		elemsForPeer.elems = append(elemsForPeer.elems, elem)
	}
	if peer.isRunning.Load() {
		peer.StagePackets(elemsForPeer)
		peer.SendStagedPackets()
	}
}

func (device *Device) SendPacketsToNetwork(packets [][]byte) {
	if len(packets) < 1 {
		return
	}

	for i := 0; i < len(packets); i++ {
		packets[i] = padForOffset(packets[i])
	}

	_, err := device.tun.device.Write(packets, MessageTransportOffsetContent)
	if err != nil && !device.isClosed() {
		device.log.Errorf("Failed to write packets to TUN device: %v", err)
	}
}
