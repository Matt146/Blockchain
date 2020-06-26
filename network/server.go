package network

import (
	"Blockchain/blockchain"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go/log"
)

// RouteIfNeeded - Should be called before doing anything with
// any packet. Routes the packet again unless if it's not directed
// to me. Returns 0 if packet is routed, returns 1 if the packet is
// for you, and returns -1 if the packet was dropped because of hop limit.
// Returns error if it errors as well, otherwise that return
// will be nil.
// Returns:
// -1: packet dropped
// 0: packed routed
// 1: packet is for me
// 2: my own packet was sent back to me
func (net *Network) RouteIfNeeded(w http.ResponseWriter, r *http.Request) (int, error) {
	// Check for correct HTTP method
	if r.Method == "GET" {
		w.Write([]byte("Fuck off. Use POST"))
		return -1, errors.New("The peer wouldn't use GET")
	}

	// Decode the HTTP form to a packet
	p, err := DeserializeFromForm(r)
	if err != nil {
		return 5000, err
	}

	// Now check destination IP and ID. If they match my destination ID
	// and IP, return because the packet is meant for me. Otherwise, keep
	// going in the function to route the packet
	if bytes.Compare(p.DestinationID, net.MyID) == 0 {
		if strings.Compare(r.FormValue("DestinationIP"), net.MyIP) == 0 {
			return 1, nil
		}
	}

	// Now, check if the source IP and ID are equal to my IP and ID. Return if it
	// is.
	if bytes.Compare(p.SourceID, net.MyID) == 0 {
		if strings.Compare(r.FormValue("SourceIP"), net.MyIP) == 0 {
			return 2, nil
		}
	}

	// Check the hop limit
	if p.HopLimit > 0 {
		p.HopLimit--
	}
	if p.HopLimit == 0 {
		return -1, nil
	}

	// Now route packet according to the correct
	// SendType
	if p.SendType == PacketSingleCast {
		err = net.SendPacket(p)
		return 0, err
	} else if p.SendType == PacketBroadCast {
		err = net.BroadcastPacket(*p)
		return 0, err
	}

	// The failsafe is just to send it
	// rather than broadcast
	err = net.SendPacket(p)
	return 0, err
}

// JoinResp - The structure of the response of a JOIN request
type JoinResp struct {
	Net *Network `json:"Network"`
	ID  []byte   `json:"ID"`
	IP  string   `json:"IP"`
}

// JoinHandler - JOIN request HTTP handler
func (net *Network) JoinHandler(w http.ResponseWriter, r *http.Request) {
	// Route packet if needed
	// Actually, don't bc with JOIN
	// you don't know the IP and stuff
	/*
		result, err := net.RouteIfNeeded(w, r)
		if err != nil {
			w.Write([]byte("Decoding error! Please try again!"))
			return
		}
	*/

	// Handle the packet now
	// Generate a response
	netID := blockchain.GenRandBytes(32)
	joinresp := &JoinResp{Net: net, ID: netID, IP: r.RemoteAddr}

	// Serialize the response
	joinRespSerialized, err := json.Marshal(joinresp)
	if err != nil {
		w.Write([]byte("JSON serialization error! Please try again!"))
		return
	}

	// Now, also add the joining node to the routing table
	net.mux.Lock()
	newNode := Node{ID: netID, IPAddr: r.RemoteAddr, CPUPower: 10, NetPower: 10}
	net.Nodes[string(netID)] = newNode
	net.mux.Unlock()

	// Now write the response over the wire
	w.Write(joinRespSerialized)
}

// LeaveHandler - LEAVE request HTTP handler
func (net *Network) LeaveHandler(w http.ResponseWriter, r *http.Request) {
	// Route packet if needed
	result, err := net.RouteIfNeeded(w, r)
	if err != nil {
		w.Write([]byte("Decoding error! Please try again!"))
		return
	}

	// Handle the packet now
	if result == 1 {
		// @TODO CHANGE: Respond with just a simple ACK
		w.Write([]byte("ACK"))

		// Now, forward the packet to everyone in the network
		p, err := DeserializeFromForm(r)
		if err != nil {
			log.Error(err)
			return
		}
		net.BroadcastPacket(*p)

		// Now and only now do we actually remove
		// the node from our routing table
		net.mux.Lock()
		delete(net.Nodes, string(p.SourceID))
		net.mux.Unlock()
	}
}
