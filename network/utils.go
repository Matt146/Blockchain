package network

import "sync"

const (
	// Version - The current version of the networking modules
	Version = 0

	// HopLimit - Default hop limit of the packet
	HopLimit = 128

	// Port - the default listening port
	Port = ":8080"
)

// Node - Models a singular node in the network
type Node struct {
	ID       []byte
	IPAddr   string // This will be the unique identifier for each node
	CPUPower int64  // The CPU power of the node
	NetPower int64  // The Network power of the node
}

// Network - represents the network
type Network struct {
	MyID  []byte          `json:"MyID"`
	MyIP  string          `json:"MyIP"`
	Nodes map[string]Node `json:"Nodes"`
	mux   sync.Mutex
}

// Packet - Models a packet for the P2P protocol.
// This should be serialized into HTTP request/response parameters
type Packet struct {
	Version       uint32
	Type          string
	SourceID      []byte
	DestinationID []byte
	SourceIP      string
	DestinationIP string
	Data          []byte // When serialized, it should be converted to base64
	HopLimit      uint32
}

// MakeNetwork - Network struct constructor
func MakeNetwork() *Network {
	return &Network{MyID: []byte(""), MyIP: "", Nodes: make(map[string]Node)}
}
