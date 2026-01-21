package device

import (
	"iter"
	"sync"
)

type Peers interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
	Del(NoisePublicKey)
	Set(NoisePublicKey, *Peer)
	Get(NoisePublicKey) (*Peer, bool)
	All() iter.Seq2[NoisePublicKey, *Peer]
	Len() int
}

type DefaultPeers struct {
	sync.RWMutex // protects keyMap
	keyMap       map[NoisePublicKey]*Peer
}

func MakeDefaultPeers() Peers {
	return &DefaultPeers{
		keyMap: make(map[NoisePublicKey]*Peer),
	}
}

func (p *DefaultPeers) All() iter.Seq2[NoisePublicKey, *Peer] {
	return func(yield func(NoisePublicKey, *Peer) bool) {
		for key, peer := range p.keyMap {
			if !yield(key, peer) {
				return
			}
		}
	}
}

func (p *DefaultPeers) Set(key NoisePublicKey, peer *Peer) {
	p.keyMap[key] = peer
}

func (p *DefaultPeers) Get(key NoisePublicKey) (*Peer, bool) {
	peer, ok := p.keyMap[key]
	return peer, ok
}

func (p *DefaultPeers) Del(key NoisePublicKey) {
	delete(p.keyMap, key)
}

func (p *DefaultPeers) Len() int {
	return len(p.keyMap)
}
