package network

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Join - This function takes a boot node
// hostname (IP:port)and joins the network.
// It gets assigned it's ID, it's IP
// and gets sent the routing table.
// In the source IP, we put the port that we
// plan to listen on with no other data.
// That way, we can use multiple ports and multiple
// devices in a NAT network.
func (net *Network) Join(addr string, listeningPort int) error {
	p := Packet{
		PVersion:      ProtocolVersion,
		Type:          "JOIN",
		SourceID:      net.MyID,
		DestinationID: []byte(""),
		SourceIP:      "" + strconv.FormatInt(int64(listeningPort), 10),
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
	req, err := http.NewRequest("POST", "http://"+addr+"/"+"JOIN", strings.NewReader(formValues.Encode()))
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
	log.Println("[+] JOIN successful!")
	for k := range net.Nodes {
		log.Printf("\t- Node received: %v\n", base64.URLEncoding.EncodeToString([]byte(k)))
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
	log.Printf("[+] Sending PING request to (%v, %v)\n", base64.URLEncoding.EncodeToString(p.DestinationID), net.Nodes[string(p.DestinationID)].IPAddr)
	err := net.SendPacket(p)
	if err != nil {
		log.Printf("[+] Error occured during PING request: %s\n", err.Error())
		return nil
	}
	log.Println("[+] PING successful!")
	return err
}

// Pong - Send this packet to respond to a ping request. If peerIP is
// empty, send the packet through the network. If it isn't, send
// the packet directly
func (net *Network) Pong(peerID []byte, peerIP string) error {
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
	log.Printf("[+] Sending PONG response (%v, %v)\n", base64.URLEncoding.EncodeToString(p.DestinationID), net.Nodes[string(p.DestinationID)].IPAddr)
	if peerIP == "" {
		err := net.SendPacket(p)
		if err != nil {
			log.Printf("[+] Error occured during PONG response: %s\n", err.Error())
			return err
		}
	}
	_, err := net.SendPacketDirectly(p)
	if err != nil {
		log.Printf("[+] Error occured during PONG response: %s\n", err.Error())
		return err
	}
	return nil
}

// SendMSG - Send a message to a peer. Sends
// a message to peerID. If you choose to send a packet directly
// without going through the flooding algorithm,
// supply a peerIP and don't just leave if blank.
// The correct server will then reply through a standard
// HTTP response, whether the packet is sent directly
// or indirectly
func (net *Network) SendMSG(peerID []byte, peerIP string, msg []byte) error {
	p := &Packet{
		PVersion:      ProtocolVersion,
		Type:          "SendMSG",
		SourceID:      net.MyID,
		DestinationID: peerID,
		SourceIP:      net.MyIP,
		DestinationIP: peerIP,
		Data:          msg,
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}
	log.Printf("[+] Sending SendMSG request (%v, %v)\n", base64.URLEncoding.EncodeToString(p.DestinationID), net.Nodes[string(p.DestinationID)].IPAddr)
	if peerIP == "" {
		err := net.SendPacket(p)
		if err != nil {
			log.Printf("[+] Error occured during SendMSG request: %s\n", err.Error())
			return err
		}
		return nil
	}
	_, err := net.SendPacketDirectly(p)
	if err != nil {
		log.Printf("[+] Error occured during SendMSG request: %s\n", err.Error())
		return err
	}
	return nil
}

// BroadcastMSG - Broadcasts a message to all peers. It will receive
// a response in the form of a BroadcastMSGResponse request, which
// it will handle through a BroadcastMSGResponseHandler
func (net *Network) BroadcastMSG(msg []byte) error {
	p := Packet{
		PVersion:      ProtocolVersion,
		Type:          "BroadcastMSG",
		SourceID:      net.MyID,
		DestinationID: []byte(""), // this gets filled in when the message gets broadcasted
		SourceIP:      net.MyIP,
		DestinationIP: "", // this gets filled in when the message gets broadcasted
		Data:          msg,
		HopLimit:      HopLimitDefault,
		SendType:      PacketBroadCast,
	}
	err := net.BroadcastPacket(p)
	return err
}

// BroadcastMSGResponse - This is a request that is sent
// to the BroadcastMSGResponseHandler in response to a BroadcastMSG.
// It can either send the packet directly or go through
// the entire flooding algorithm to get the packet to the correct computer.
func (net *Network) BroadcastMSGResponse(peerID []byte, peerIP string, msg []byte) error {
	p := &Packet{
		PVersion:      ProtocolVersion,
		Type:          "BroadcastMSGResponse",
		SourceID:      net.MyID,
		DestinationID: peerID, // this gets filled in when the message gets broadcasted
		SourceIP:      net.MyIP,
		DestinationIP: peerIP, // this gets filled in when the message gets broadcasted
		Data:          msg,
		HopLimit:      HopLimitDefault,
		SendType:      PacketSingleCast,
	}
	log.Printf("[+] Sending BroadcastMSGResponse response (%v, %v)\n", base64.URLEncoding.EncodeToString(p.DestinationID), net.Nodes[string(p.DestinationID)].IPAddr)
	if peerIP == "" {
		err := net.SendPacket(p)
		if err != nil {
			log.Printf("[+] Error occured during BroadcastMSGResponse response: %s\n", err.Error())
			return err
		}
	}
	_, err := net.SendPacketDirectly(p)
	if err != nil {
		log.Printf("[+] Error occured during BroadcastMSGResponse response: %s\n", err.Error())
		return err
	}
	return nil
}
