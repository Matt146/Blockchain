package network

import (
	"Blockchain/blockchain"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		return -1, errors.New("RouteIfNeeded: Invalid request method: GET")
	}

	// Decode the HTTP form to a packet
	p, err := DeserializeFromForm(r)
	if err != nil {
		fmt.Println("DESERIALIZATION ERROR")
		fmt.Println(err)
		return 5000, err
	}

	// Print debug information about packet
	fmt.Println("Packet DEBUG information:")
	fmt.Println("=========================")
	fmt.Printf("Source ID: %v\n", p.SourceID)
	fmt.Printf("Destination ID: %v\n", p.DestinationID)
	fmt.Printf("My ID: %v\n", net.MyID)
	fmt.Printf("Source IP: %v\n", p.SourceIP)
	fmt.Printf("Destination IP: %v\n", p.DestinationIP)
	fmt.Printf("My IP: %v\n", p.DestinationIP)
	fmt.Printf("Type: %v\n", p.Type)
	fmt.Println("=========================")

	/*
		// Now check destination IP and ID. If they match my destination ID
		// and IP, return because the packet is meant for me. Otherwise, keep
		// going in the function to route the packet
		if bytes.Compare(p.DestinationID, net.MyID) == 0 {
			if strings.Compare(r.FormValue("DestinationIP"), net.MyIP) == 0 {
				fmt.Printf("Packet is for me :D\n")
				return 1, nil
			}
		}
	*/

	// Now check the destination ID. If it matches my ID, return because
	// the packet is meant for me. Otherwise, keep going in the function
	// to route the packet
	if bytes.Compare(p.DestinationID, net.MyID) == 0 {
		fmt.Println("Packet is for me: :D")
		return 1, nil
	}

	// Now, check if the source IP and ID are equal to my IP and ID. Return if it
	// is.
	if bytes.Compare(p.SourceID, net.MyID) == 0 {
		if strings.Compare(r.FormValue("SourceIP"), net.MyIP) == 0 {
			fmt.Printf("My own packet was sent back to me :(\n")
			fmt.Printf("ID: Source: %v | Destination: %v\n", p.SourceID, p.DestinationID)
			fmt.Printf("IP: Source: %v | Destination: %v\n", p.SourceIP, p.DestinationIP)
			fmt.Printf("My ID: %v\n", net.MyID)
			fmt.Printf("My IP: %v\n", net.MyIP)
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
		fmt.Println("SEND PACKET #1")
		return 0, err
	} else if p.SendType == PacketBroadCast {
		err = net.BroadcastPacket(*p)
		fmt.Println("SEND PACKET #2")
		return 0, err
	}

	// The failsafe is just to send it
	// rather than broadcast
	err = net.SendPacket(p)
	fmt.Println("SEND PACKET #3")
	return 0, err
}

// JoinResp - The structure of the response of a JOIN request
type JoinResp struct {
	Net           *Network `json:"Network"`
	ID            []byte   `json:"ID"`
	IP            string   `json:"IP"`
	BootstrapNode Node     `json:"BootstrapNode"`
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
	ipParsed := strings.Split(r.RemoteAddr, ":")[0]
	ipParsed += ":" + r.FormValue("SourceIP")
	joinresp := &JoinResp{Net: net, ID: netID, IP: ipParsed,
		BootstrapNode: Node{
			ID:       net.MyID,
			IPAddr:   net.MyIP,
			CPUPower: 10,
			NetPower: 10,
		}}

	// Serialize the response
	joinRespSerialized, err := json.Marshal(joinresp)
	if err != nil {
		w.Write([]byte("JSON serialization error! Please try again!"))
		return
	}

	fmt.Println("\tJOIN RESPONSE JSON:")
	fmt.Printf("===============================\n")
	fmt.Printf("%v\n", string(joinRespSerialized))
	fmt.Printf("===============================\n")

	// Now, also add the joining node to the routing table
	net.mux.Lock()
	newNode := Node{ID: netID, IPAddr: ipParsed, CPUPower: 10, NetPower: 10}
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

// PingHandler - Handle a ping packet
func (net *Network) PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received ping!")
	result, err := net.RouteIfNeeded(w, r)
	if err != nil {
		fmt.Println("ERROR HAPPENED")
		fmt.Println(err)
		w.Write([]byte("Decoding error! Please try again!"))
		return
	}

	fmt.Printf("Packet router output: %d\n", result)
	if result == 1 {
		fmt.Println("PROCESSING PING")
		// @TODO CHANGE: Respond with just a simple ACK
		w.Write([]byte("ACK"))

		// Now, deserialize the packet to get the source IP
		pRecved, err := DeserializeFromForm(r)
		if err != nil {
			log.Error(err)
			return
		}

		// Now, send a pong response
		err = net.Pong(pRecved.SourceID, pRecved.SourceIP)
		log.Error(err)
	}
}

// PongHandler - Handle a pong packet
func (net *Network) PongHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PONG :D")
	result, err := net.RouteIfNeeded(w, r)
	if err != nil {
		fmt.Println(result)
		w.Write([]byte("Decoding error! Please try again!"))
		return
	}

	if result == 1 {
		fmt.Println("HANDLING PONG! YAY!")
		// @TODO CHANGE: Respond with just a simple ACK
		w.Write([]byte("ACK"))

		// Now, deserialize the packet to get the source IP
		// and source ID
		pRecved, err := DeserializeFromForm(r)
		if err != nil {
			log.Error(err)
			return
		}

		// Now, cache the node
		net.mux.Lock()
		net.Nodes[string(pRecved.SourceID)] = Node{
			ID:       pRecved.SourceID,
			IPAddr:   pRecved.SourceIP,
			CPUPower: 10,
			NetPower: 10,
		}
		net.mux.Unlock()
	}
}
