package main

import (
	"Blockchain/blockchain"
	"Blockchain/network"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

var wg sync.WaitGroup

var servers []string = []string{
	"127.0.0.1:" + network.Port,
}

func main() {
	/*Test mining*/
	fmt.Println("[+] Testing blockchain...")
	b := blockchain.Block{
		Index:      1,
		Timestamp:  uint64(time.Now().Unix()), // if time machines are a thing, this code is broken
		Difficulty: 1,
	}
	b.MineBlock()

	/*Test networking*/
	fmt.Println("[+] Testing network...")

	// Start first node in the network
	net := network.MakeNetwork()
	network.InitMSGQueue()
	// Wont' work because it uses public IP and my router has NAT: net.BootstrapNetwork()
	net.MyIP = "127.0.0.1" + network.Port
	net.MyID = blockchain.GenRandBytes(32)
	http.HandleFunc("/JOIN", net.JoinHandler)
	http.HandleFunc("/PING", net.PingHandler)
	http.HandleFunc("/PONG", net.PongHandler)
	http.HandleFunc("/SendMSG", net.SendMSGHandler)
	http.HandleFunc("/BroadcastMSG", net.BroadcastMSGHandler)
	http.HandleFunc("/BroadcastMSGResponse", net.BroadcastMSGResponseHandler)
	wg.Add(1)
	defer wg.Done()
	go func() {
		log.Error(http.ListenAndServe(network.Port, nil))
	}()

	fmt.Println("Client sending JOIN request")

	// Just echo the messages now
	for {
		for k := range net.Nodes {
			packets := network.HandleMsgQueuePackets(net.Nodes[k].ID)
			for _, v := range packets {
				fmt.Println(string(v.Data))
			}
		}
	}

	// Wait now
	wg.Wait()
}
