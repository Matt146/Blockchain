package network

import (
	"sync"
)

var mux sync.Mutex

// msgQueue - Map between the SourceID of the packet to a slice
// of msg packets from that SourceID
var msgQueue map[string][]Packet

// InitMSGQueue - Initializes the msgQueue. This needs to be called before using
// the message queue
func InitMSGQueue() {
	msgQueue = make(map[string][]Packet)
}

// AddToMsgQueue - Adds a msg packet to the msgQueue
// in a thread-safe way
func (p *Packet) AddToMsgQueue() {
	mux.Lock()
	defer mux.Unlock()
	msgQueue[string(p.SourceID)] = append(msgQueue[string(p.SourceID)], *p)
}

// HandleMsgQueuePackets - returns a copy of all
// the pending packets in the queue for a specific peer to be handled
// and then deletes all the packets in the queue for that peer.
// Blocks the entire queue while running
func HandleMsgQueuePackets(peerID []byte) []Packet {
	mux.Lock()
	defer mux.Unlock()
	packets := msgQueue[string(peerID)]
	delete(msgQueue, string(peerID))
	return packets
}
