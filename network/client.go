package network

import (
	"Blockchain/blockchain"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Join - This function takes a boot node
// hostname and returns a packet that
// would be used to send a request to the network
func (net *Network) Join(addr string) Packet {
	p := Packet{
		Version:       Version,
		Type:          "JOIN",
		SourceID:      net.MyID,
		DestinationID: []byte(""),
		SourceIP:      "",
		DestinationIP: addr,
		Data:          []byte("Pls man. Let me join"),
		HopLimit:      HopLimitDefault,
	}

	return p
}

// SendPacket - Use this function to send packets
// to servers
func (net *Network) SendPacket(p *Packet) error {
	// Create the client and the request
	client := &http.Client{}

	// Craft form values for request
	formValues := url.Values{
		"Version":       {strconv.FormatUint(uint64(p.Version), blockchain.Base)},
		"Type":          {p.Type},
		"SourceID":      {base64.URLEncoding.EncodeToString(p.SourceID)},
		"DestinationID": {base64.URLEncoding.EncodeToString(p.DestinationID)},
		"SourceIP":      {p.SourceIP},
		"DestinationIP": {p.DestinationIP},
		"Data":          {base64.URLEncoding.EncodeToString(p.Data)},
		"HopLimit":      {strconv.FormatUint(uint64(p.HopLimit), blockchain.Base)},
	}

	// Loop through and broadcast the packet
	for i := range net.Nodes {
		// Craft the request
		req, err := http.NewRequest("POST", net.Nodes[i].IPAddr+Port+"/"+p.Type, strings.NewReader(formValues.Encode()))
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
