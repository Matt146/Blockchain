package network

import (
	"Blockchain/blockchain"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const (
	// ProtocolVersion - The current version of the networking modules
	ProtocolVersion = 1

	// HopLimitDefault - Default hop limit of the packet
	HopLimitDefault = 128

	// Port - the default listening port
	Port = ":8080"
)

const (
	// PacketSingleCast - The SendType for a singlecast packet
	PacketSingleCast = 0

	// PacketBroadCast - The SendType for a broacast packet
	// (meaning the packet is intended for all nodes)
	// on the network
	PacketBroadCast = 1
)

// Node - Models a singular node in the network
type Node struct {
	ID       []byte `json:"ID"`       // Unique identifier for each node
	IPAddr   string `json:"IPAddr"`   // IP:Port
	CPUPower int64  `json:"CPUPower"` // The CPU power of the node
	NetPower int64  `json:"NetPower"` // The Network power of the node
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
	PVersion      int32
	Type          string
	SourceID      []byte
	DestinationID []byte
	SourceIP      string
	DestinationIP string
	Data          []byte // When serialized, it should be converted to base64
	HopLimit      int32
	SendType      int32
}

// getMyIP - Uses the ipify API to get my IP
// Returns an empty string if an error occurs
func getMyIP() string {
	// Create the Client
	client := &http.Client{}

	// Create a request now
	req, err := http.NewRequest("GET", "https://api.ipify.org/", nil)
	if err != nil {
		return ""
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}

	// Read the request now
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ""
	}

	return string(data)
}

// MakeNetwork - Network struct constructor
func MakeNetwork() *Network {
	return &Network{MyID: []byte(""), MyIP: "", Nodes: make(map[string]Node)}
}

// BootstrapNetwork - Call this function after MakeNetwork if
// you are the first node in the network
func (net *Network) BootstrapNetwork() {
	// Get my IP
	myIP := getMyIP()
	if myIP == "" {
		fmt.Println("Unable to get my IP")
	}
	fmt.Printf("My IP: %s\n", myIP)

	// Generate a GUID
	myID := blockchain.GenRandBytes(32)

	// Assign GUID and IP to my data
	net.MyID = myID
	net.MyIP = myIP
}

// SerializeToForm - Serialized a packet to an HTTP form
func (p *Packet) SerializeToForm() url.Values {
	fmt.Printf("Protocol version given to packet serializer: %d\n", p.PVersion)
	fmt.Printf("Serialized version given to packet serializer: %s\n", strconv.FormatInt(int64(p.PVersion), blockchain.Base))
	fmt.Printf("Destination ID given to packet serialzier: %v\n", p.DestinationID)
	fmt.Printf("Serialzied destination ID given to packet serialzier: %s\n", base64.URLEncoding.EncodeToString(p.DestinationID))
	formValues := url.Values{
		"PVersion":      {strconv.FormatInt(int64(p.PVersion), blockchain.Base)},
		"Type":          {p.Type},
		"SourceID":      {base64.URLEncoding.EncodeToString(p.SourceID)},
		"DestinationID": {base64.URLEncoding.EncodeToString(p.DestinationID)},
		"SourceIP":      {p.SourceIP},
		"DestinationIP": {p.DestinationIP},
		"Data":          {base64.URLEncoding.EncodeToString(p.Data)},
		"HopLimit":      {strconv.FormatInt(int64(p.HopLimit), blockchain.Base)},
		"SendType":      {strconv.FormatInt(int64(p.SendType), blockchain.Base)},
	}

	return formValues
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
	version, err := strconv.ParseInt(r.FormValue("PVersion"), blockchain.Base, 32)
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
	hoplimit, err := strconv.ParseInt(r.FormValue("HopLimit"), blockchain.Base, 32)
	if err != nil {
		return nil, err
	}
	sendtype, err := strconv.ParseInt(r.FormValue("SendType"), blockchain.Base, 32)
	if err != nil {
		// this is optional and defaults to singlecast
		sendtype = int64(PacketSingleCast)
	}

	// Now, take the parsed data from the form and
	// put it into a packet, which you can send off to
	// other nodes in the network
	p := &Packet{
		PVersion:      int32(version),
		Type:          type1,
		SourceID:      sourceIDDecoded,
		DestinationID: destinationIDDecoded,
		SourceIP:      sourceip,
		DestinationIP: destinationip,
		Data:          data,
		HopLimit:      int32(hoplimit),
		SendType:      int32(sendtype),
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

	// Debug Send SendPacket
	fmt.Println("SendPacket DEBUG information:")
	fmt.Println("=============================")
	fmt.Printf("Source ID: %v\n", p.SourceID)
	fmt.Printf("Destination ID: %v\n", p.DestinationID)
	fmt.Printf("My ID: %v\n", net.MyID)
	fmt.Printf("Source IP: %v\n", p.SourceIP)
	fmt.Printf("Destination IP: %v\n", p.DestinationIP)
	fmt.Printf("My IP: %v\n", p.DestinationIP)
	fmt.Printf("Type: %v\n", p.Type)
	fmt.Println("=============================")

	// Loop through and broadcast the packet
	for i := range net.Nodes {
		// Craft the request
		req, err := http.NewRequest("POST", "http://"+net.Nodes[i].IPAddr+"/"+p.Type, strings.NewReader(formValues.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formValues.Encode())))
		fmt.Printf("URL: %s://%s%s\n", req.URL.Scheme, req.URL.Host, req.URL.Path)
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

// SendPacketDirectly - Use this function to send
// a packet directly to an IP address without having
// to deal with going through the entire network.
// The IP addresss to send it to is specified in the p.DestinationIP
func (net *Network) SendPacketDirectly(p *Packet) (*http.Response, error) {
	// Create the client
	client := &http.Client{}

	// Craft the form values for the request
	formValues := p.SerializeToForm()

	// Debug Send SendPacketDirectly
	fmt.Println("SendPacketDirectly DEBUG information:")
	fmt.Println("=============================")
	fmt.Printf("Source ID: %v\n", p.SourceID)
	fmt.Printf("Destination ID: %v\n", p.DestinationID)
	fmt.Printf("My ID: %v\n", net.MyID)
	fmt.Printf("Source IP: %v\n", p.SourceIP)
	fmt.Printf("Destination IP: %v\n", p.DestinationIP)
	fmt.Printf("My IP: %v\n", p.DestinationIP)
	fmt.Printf("Type: %v\n", p.Type)
	fmt.Println("=============================")

	// Craft the request
	req, err := http.NewRequest("POST", "http://"+p.DestinationIP+"/"+p.Type, strings.NewReader(formValues.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formValues.Encode())))
	fmt.Printf("URL: %s://%s%s\n", req.URL.Scheme, req.URL.Host, req.URL.Path)
	if err != nil {
		return nil, err
	}

	// Send the request
	resp, err := client.Do(req)
	return resp, err
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
		p.SendType = PacketBroadCast

		// Craft form values for request
		formValues := p.SerializeToForm()

		// Debug Send BroadcastPacket
		fmt.Println("BroadcastPacket DEBUG information:")
		fmt.Println("==================================")
		fmt.Printf("Source ID: %v\n", p.SourceID)
		fmt.Printf("Destination ID: %v\n", p.DestinationID)
		fmt.Printf("My ID: %v\n", net.MyID)
		fmt.Printf("Source IP: %v\n", p.SourceIP)
		fmt.Printf("Destination IP: %v\n", p.DestinationIP)
		fmt.Printf("My IP: %v\n", p.DestinationIP)
		fmt.Printf("Type: %v\n", p.Type)
		fmt.Println("==================================")

		// Craft the request
		req, err := http.NewRequest("POST", "http://"+net.Nodes[i].IPAddr+"/"+p.Type, strings.NewReader(formValues.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(formValues.Encode())))
		fmt.Printf("URL: %s://%s%s\n", req.URL.Scheme, req.URL.Host, req.URL.Path)
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
