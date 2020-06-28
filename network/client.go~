package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Join - This function takes a boot node
// hostname and joins the network.
// It gets assigned it's ID, it's IP
// and gets sent the routing table.
func (net *Network) Join(addr string) error {
	p := Packet{
		PVersion:      ProtocolVersion,
		Type:          "JOIN",
		SourceID:      net.MyID,
		DestinationID: []byte(""),
		SourceIP:      "",
		DestinationIP: addr,
		Data:          []byte("Pls man. Let me join"),
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}

	// Create the client and the request
	client := &http.Client{}

	// Craft form values for request. Ideally, we would use a packet
	// and serializing that into form values but idgaf
	formValues := p.SerializeToForm()

	// Craft the request
	req, err := http.NewRequest("POST", "http://"+addr+Port+"/"+"JOIN", strings.NewReader(formValues.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formValues.Encode())))
	if err != nil {
		return err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read the raw response data
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	// Deserialize the response
	var joinRespDeserialized JoinResp
	err = json.Unmarshal(respBody, &joinRespDeserialized)
	if err != nil {
		return err
	}

	// Then, take the deserialized response and set it to the network
	net.mux.Lock()
	net.Nodes = joinRespDeserialized.Net.Nodes
	net.Nodes[string(joinRespDeserialized.BootstrapNode.ID)] = joinRespDeserialized.BootstrapNode
	net.MyID = joinRespDeserialized.ID
	net.MyIP = joinRespDeserialized.IP
	net.mux.Unlock()

	// Now, print some debug information:
	for k := range net.Nodes {
		fmt.Printf("[+] Node received: %v\n", []byte(k))
	}

	return nil
}

// Ping - Used to ping another peer to make sure that they are still active
func (net *Network) Ping(peerID []byte) error {
	p := &Packet{
		PVersion:      ProtocolVersion,
		Type:          "PING",
		SourceID:      net.MyID,
		DestinationID: peerID,
		SourceIP:      net.MyIP,
		DestinationIP: "",
		Data:          []byte("Ping :D"),
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}
	err := net.SendPacket(p)
	return err
}

// Pong - Send this packet to respond to a ping request. If peerIP is
// empty, send the packet through the network. If it isn't, send
// the packet directly
func (net *Network) Pong(peerID []byte, peerIP string) error {
	fmt.Println("Sending PONG")
	p := &Packet{
		PVersion:      ProtocolVersion,
		Type:          "PONG",
		SourceID:      net.MyID,
		DestinationID: peerID,
		SourceIP:      net.MyIP,
		DestinationIP: peerIP,
		Data:          []byte("Pong :D"),
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}
	if peerIP == "" {
		err := net.SendPacket(p)
		return err
	}
	_, err := net.SendPacketDirectly(p)
	return err
}

// SendMSG - Send a message to a peer
func (net *Network) SendMSG(peerID []byte) error {
	return nil
}

// BroadcastMSG - Broadcasts a message to all peers
func (net *Network) BroadcastMSG(data []byte) error {
	p := Packet{
		PVersion:      ProtocolVersion,
		Type:          "MSG",
		SourceID:      net.MyID,
		DestinationID: []byte(""), // this gets filled in when the message gets broadcasted
		SourceIP:      net.MyIP,
		DestinationIP: "", // this gets filled in when the message gets broadcasted
		Data:          data,
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}
	err := net.BroadcastPacket(p)
	return err
}
