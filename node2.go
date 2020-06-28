package main

import (
	"Blockchain/network"
	"fmt"
	"net/http"

	"sync"

	"github.com/opentracing/opentracing-go/log"
)

var wg sync.WaitGroup

func main() {
	// Start second node in the network
	net1 := network.MakeNetwork()
	http.HandleFunc("/JOIN", net1.JoinHandler)
	http.HandleFunc("/PING", net1.PingHandler)
	http.HandleFunc("/PONG", net1.PongHandler)
	wg.Add(1)
	defer wg.Done()
	go func() {
		log.Error(http.ListenAndServe(":6700", nil))
	}()
	err := net1.Join("127.0.0.1"+network.Port, 6700)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ping the first client and any other clients that the
	// second node may be connected to
	for k := range net1.Nodes {
		fmt.Println("[+] Pinging node: %s", net1.Nodes[k].ID)
		err := net1.Ping(net1.Nodes[k].ID)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Wait now
	wg.Wait()
}
