package main

import (
	"Blockchain/network"
	"fmt"
	"net/http"
	"strconv"
	"time"

	mrand "math/rand"
	"sync"

	"github.com/opentracing/opentracing-go/log"
)

var wg sync.WaitGroup

func main() {
	// Generate a random port number
	mrand.Seed(time.Now().UTC().UnixNano())
	randPort := mrand.Intn(45000) + 10000

	// Start second node in the network
	net1 := network.MakeNetwork()
	network.InitMSGQueue()
	http.HandleFunc("/JOIN", net1.JoinHandler)
	http.HandleFunc("/PING", net1.PingHandler)
	http.HandleFunc("/PONG", net1.PongHandler)
	http.HandleFunc("/SendMSG", net1.SendMSGHandler)
	http.HandleFunc("/BroadcastMSG", net1.BroadcastMSGHandler)
	http.HandleFunc("/BroadcastMSGResponse", net1.BroadcastMSGResponseHandler)
	wg.Add(1)
	defer wg.Done()
	go func() {
		log.Error(http.ListenAndServe(":"+strconv.FormatInt(int64(randPort), 10), nil))
	}()
	err := net1.Join("127.0.0.1"+network.Port, randPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ping the first client and any other clients that the
	// second node may be connected to
	for k := range net1.Nodes {
		fmt.Printf("[+] Pinging node: %s\n", net1.Nodes[k].ID)
		err := net1.Ping(net1.Nodes[k].ID)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("[+] SendMSG to node: %s\n", net1.Nodes[k].ID)
		err = net1.SendMSG(net1.Nodes[k].ID, "", []byte("HELLO! HOW ARE YOU DOING RN?"))
	}

	// Wait now
	wg.Wait()
}
