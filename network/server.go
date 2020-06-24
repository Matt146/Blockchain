package network

import (
	"Blockchain/blockchain"
	"bytes"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
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
	}

	// Now decode destination ID
	destinationIDDecoded, err := base64.URLEncoding.DecodeString(r.FormValue("DestinationID"))
	if err != nil {
		return 5000, err
	}

	// Now check destination IP and ID. If they match my destination ID
	// and IP, return because the packet is meant for me. Otherwise, keep
	// going in the function to route the packet
	if bytes.Compare(destinationIDDecoded, net.MyID) == 0 {
		if strings.Compare(r.FormValue("DestinationIP"), net.MyIP) == 0 {
			return 1, nil
		}
	}

	// Now decode source ID
	sourceIDDecoded, err := base64.URLEncoding.DecodeString(r.FormValue("SourceID"))
	if err != nil {
		return 5000, err
	}

	// Now, check if the source IP and ID are equal to my IP and ID. Return if it
	// is.
	if bytes.Compare(sourceIDDecoded, net.MyID) == 0 {
		if strings.Compare(r.FormValue("SourceIP"), net.MyIP) == 0 {
			return 2, nil
		}
	}

	// Now, parse the data out of the HTTP POST parameters
	version, err := strconv.ParseUint(r.FormValue("Version"), blockchain.Base, 32)
	if err != nil {
		return 5000, nil
	}
	type1 := r.FormValue("Type")
	sourceid, err := base64.URLEncoding.DecodeString(r.FormValue("SourceID"))
	if err != nil {
		return 5000, nil
	}
	destinationid, err := base64.URLEncoding.DecodeString(r.FormValue("DestinationIP"))
	if err != nil {
		return 5000, nil
	}
	data, err := base64.URLEncoding.DecodeString(r.FormValue("Data"))
	if err != nil {
		return 5000, nil
	}
	sourceip := r.FormValue("SourceIP")
	destinationip := r.FormValue("DestinationIP")
	hoplimit, err := strconv.ParseUint(r.FormValue("HopLimit"), blockchain.Base, 32)
	if err != nil {
		return 5000, nil
	}

	// Now, take the parsed data from the form and
	// put it into a packet, which you can send off to
	// other nodes in the network
	p := Packet{
		Version:       uint32(version),
		Type:          type1,
		SourceID:      sourceid,
		DestinationID: destinationid,
		SourceIP:      sourceip,
		DestinationIP: destinationip,
		Data:          data,
		HopLimit:      uint32(hoplimit),
	}
	if p.HopLimit > 0 {
		p.HopLimit--
	}
	if p.HopLimit == 0 {
		return -1, nil
	}

	// Now route packet
	net.SendPacket(&p)
	return 0, nil
}

// JoinHandler - JOIN request HTTP handler
func (net *Network) JoinHandler(w http.ResponseWriter, r *http.Request) {
	// Route packet if needed
	result, err := net.RouteIfNeeded(w, r)
	if err != nil {
		w.Write([]byte("Decoding error! Please try again!"))
	}

	// Handle the packet now
	if result == 1 {
		//netID := blockchain.GenRandBytes(32)

	}
}
