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

	// Start Server
	net := network.MakeNetwork()
	http.HandleFunc("/JOIN", net.JoinHandler)
	http.HandleFunc("/PING", net.PingHandler)
	wg.Add(1)
	defer wg.Done()
	go func() {
		log.Error(http.ListenAndServe(network.Port, nil))
	}()

	fmt.Println("Client sending JOIN request")

	// Start client
	err := net.Join("localhost")
	if err != nil {
		log.Error(err)
	}
	fmt.Printf("Client ID: %s\n", net.MyIP)
	for k := range net.Nodes {
		net.Ping([]byte(k))
	}

	// Wait now
	wg.Wait()
}
