package network

import (
	"Blockchain/blockchain"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	// Version - The current version of the networking modules
	Version = 0

	// HopLimitDefault - Default hop limit of the packet
	HopLimitDefault = 128

	// Port - the default listening port
	Port = ":8080"
)

const (
	// PacketSingleCast - The SendType for a singlecast packet
	PacketSingleCast = uint32(0)

	// PacketBroadCast - The SendType for a broacast packet
	// (meaning the packet is intended for all nodes)
	// on the network
	PacketBroadCast = uint32(1)
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
// SendType:
// 0 - singlecast
// 1 - broadcast
type Packet struct {
	Version       uint32
	Type          string
	SourceID      []byte
	DestinationID []byte
	SourceIP      string
	DestinationIP string
	Data          []byte // When serialized, it should be converted to base64
	HopLimit      uint32
	SendType      uint32
}

// MakeNetwork - Network struct constructor
func MakeNetwork() *Network {
	return &Network{MyID: []byte(""), MyIP: "", Nodes: make(map[string]Node)}
}

// SerializeToForm - Serialized a packet to an HTTP form
func (p *Packet) SerializeToForm() url.Values {
	return url.Values{
		"Version":       {strconv.FormatUint(uint64(p.Version), blockchain.Base)},
		"Type":          {p.Type},
		"SourceID":      {base64.URLEncoding.EncodeToString(p.SourceID)},
		"DestinationID": {base64.URLEncoding.EncodeToString(p.DestinationID)},
		"SourceIP":      {p.SourceIP},
		"DestinationIP": {p.DestinationIP},
		"Data":          {base64.URLEncoding.EncodeToString(p.Data)},
		"HopLimit":      {strconv.FormatUint(uint64(p.HopLimit), blockchain.Base)},
		"SendType":      {strconv.FormatUint(uint64(p.SendType), blockchain.Base)},
	}
}

// DeserializeFromForm - converts HTTP form representation of
// packet to actual packet struct
func DeserializeFromForm(r *http.Request) (*Packet, error) {
	// Now decode destination ID
	destinationIDDecoded, err := base64.URLEncoding.DecodeString(r.FormValue("DestinationID"))
	if err != nil {
		return nil, err
	}

	// Now decode source ID
	sourceIDDecoded, err := base64.URLEncoding.DecodeString(r.FormValue("SourceID"))
	if err != nil {
		return nil, err
	}

	// Now, parse the rest of the data out of the HTTP POST parameters
	version, err := strconv.ParseUint(r.FormValue("Version"), blockchain.Base, 32)
	if err != nil {
		return nil, err
	}
	type1 := r.FormValue("Type")
	data, err := base64.URLEncoding.DecodeString(r.FormValue("Data"))
	if err != nil {
		return nil, err
	}
	sourceip := r.FormValue("SourceIP")
	destinationip := r.FormValue("DestinationIP")
	hoplimit, err := strconv.ParseUint(r.FormValue("HopLimit"), blockchain.Base, 32)
	if err != nil {
		return nil, err
	}
	sendtype, err := strconv.ParseUint(r.FormValue("SendType"), blockchain.Base, 32)
	if err != nil {
		// this is optional and defaults to singlecast
		sendtype = uint64(PacketSingleCast)
	}

	// Now, take the parsed data from the form and
	// put it into a packet, which you can send off to
	// other nodes in the network
	p := &Packet{
		Version:       uint32(version),
		Type:          type1,
		SourceID:      sourceIDDecoded,
		DestinationID: destinationIDDecoded,
		SourceIP:      sourceip,
		DestinationIP: destinationip,
		Data:          data,
		HopLimit:      uint32(hoplimit),
		SendType:      uint32(sendtype),
	}

	return p, nil
}

// SendPacket - Use this function to send packets
// to a specific server through a flooding algorithm
func (net *Network) SendPacket(p *Packet) error {
	// Create the client and the request
	client := &http.Client{}

	// Craft form values for request
	formValues := p.SerializeToForm()

	// Loop through and broadcast the packet
	for i := range net.Nodes {
		// Craft the request
		req, err := http.NewRequest("POST", "http://"+net.Nodes[i].IPAddr+Port+"/"+p.Type, strings.NewReader(formValues.Encode()))
		if err != nil {
			return err
		}

		// Send the request
		_, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	// Return now
	return nil
}

// BroadcastPacket - Use this function to
// broadcast a packet to all other nodes in
// a network
func (net *Network) BroadcastPacket(p Packet) error {
	// Create the client and the request
	client := &http.Client{}

	// Loop through and broadcast the packet
	for i := range net.Nodes {
		// Change the correct parameters of
		// the packet
		p.DestinationID = net.Nodes[i].ID
		p.DestinationIP = net.Nodes[i].IPAddr
		p.SendType = PacketSingleCast

		// Craft form values for request
		formValues := p.SerializeToForm()

		// Craft the request
		req, err := http.NewRequest("POST", "http://"+net.Nodes[i].IPAddr+Port+"/"+p.Type, strings.NewReader(formValues.Encode()))
		if err != nil {
			return err
		}

		// Send the request
		_, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	// Return now
	return nil
}
