package device

// FIXME: Document how this interacts with the rest of wireguard

import ()

type PacketDirection int

const (
	PacketDrop PacketDirection = iota
	PacketFromTun
	PacketFromPeer
	PacketToTun
	PacketToPeer
	PacketNone PacketDirection = 9999
)

type PacketFilterPacket struct {
	Packet []byte
	Dir    PacketDirection
	Peer   *Peer
}

type PacketFilter interface {
	FilterFromTun([]byte, int64) []byte
	FilterToTun([]byte, *Peer, int64) []byte
	GetExtraPackets() []PacketFilterPacket
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

func (pf *DefaultPacketFilter) GetExtraPackets() []PacketFilterPacket {
	return []PacketFilterPacket{}
}

func (device *Device) processQeueudExtraPackets(pf PacketFilter, maxBatchSize int) {
	device.processExtraPackets(pf, nil, maxBatchSize)
}

func (device *Device) processExtraPackets(pf PacketFilter, elemsByPeer map[*Peer]*QueueOutboundElementsContainer, maxBatchSize int) {
	device.extraPacketsLock.Lock()
	defer device.extraPacketsLock.Unlock()

	// Allocate buffers for things we will send directly
	toTunBufs := make([][]byte, 10)
	toTunBufs = toTunBufs[:0]

	// Check for extra packets we need to handle
	device.extraPackets = append(device.extraPackets, pf.GetExtraPackets()...)

	// Iterate over device.extraPackets, removing any we know how
	// to handle right now, leaving the rest.
	devExtras := device.extraPackets[:0]
	for i := range device.extraPackets {
		extra := device.extraPackets[i]
		if extra.Dir == PacketToTun && len(toTunBufs) < maxBatchSize {
			toTunBufs = append(toTunBufs, padForOffset(extra.Packet))
		} else if extra.Dir == PacketToPeer && extra.Peer != nil && elemsByPeer != nil {
			elemsForPeer, ok := elemsByPeer[extra.Peer]
			if !ok {
				elemsForPeer = device.GetOutboundElementsContainer()
				elemsByPeer[extra.Peer] = elemsForPeer
			}
			// Sending to the peer, we do not need to concern
			// ourselves with elem.buffer, the packet is enough.
			elem := device.NewOutboundElement()
			elem.packet = extra.Packet
			elemsForPeer.elems = append(elemsForPeer.elems, elem)
			device.log.Verbosef("Added to peer %+v: %v", extra.Peer, elem)
		} else {
			devExtras = append(devExtras, extra)
		}
	}

	if len(toTunBufs) > 0 {
		_, err := device.tun.device.Write(toTunBufs, MessageTransportOffsetContent)
		if err != nil && !device.isClosed() {
			device.log.Errorf("Failed to write packets to TUN device: %v", err)
		}
	}

	device.extraPackets = devExtras
}
