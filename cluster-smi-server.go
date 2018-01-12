package main

import (
	"github.com/patwie/cluster-smi/cluster"
	"github.com/patwie/cluster-smi/messaging"
	"github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack"
	"log"
	"sync"
)

// nice cluster struct
var clus cluster.Cluster

// intermediate struct (under mutex lock)
var allNodes map[string]cluster.Node

func main() {

	// load ports and ip-address
	cfg := CreateConfig()

	allNodes = make(map[string]cluster.Node)
	var mutex = &sync.Mutex{}

	// message loop
	log.Println("Cluster-SMI-Server is active. Press CTRL+C to shut down.")

	// receiving messages in extra thread
	go func() {
		// incoming messages (PUSH-PULL)
		SocketAddr := "tcp://" + "*" + ":" + cfg.ServerPortGather
		log.Println("Now listening on", SocketAddr)
		node_socket, err := zmq4.NewSocket(zmq4.PULL)
		if err != nil {
			panic(err)
		}
		defer node_socket.Close()
		node_socket.Bind(SocketAddr)

		for {
			// read node information
			s, err := node_socket.RecvBytes(0)
			if err != nil {
				log.Println(err)
				continue
			}

			var node cluster.Node
			err = msgpack.Unmarshal(s, &node)
			if err != nil {
				log.Println(err)
				continue
			}

			mutex.Lock()
			allNodes[node.Name] = node
			mutex.Unlock()

		}
	}()

	// outgoing messages (REQ-ROUTER)
	SocketAddr := "tcp://" + "*" + ":" + cfg.ServerPortDistribute
	log.Println("Router binds to", SocketAddr)
	router_socket, err := zmq4.NewSocket(zmq4.ROUTER)
	if err != nil {
		panic(err)
	}
	defer router_socket.Close()
	router_socket.Bind(SocketAddr)

	for {

		// read request of client
		msg, err := messaging.ReceiveMultipartMessage(router_socket)
		if err != nil {
			panic(err)
		}

		mutex.Lock()
		// rebuild cluster struct from map
		clus := cluster.Cluster{}
		for _, n := range allNodes {
			clus.Nodes = append(clus.Nodes, n)
		}
		mutex.Unlock()

		// send cluster information to client
		body, err := msgpack.Marshal(&clus)
		msg.Body = body
		messaging.SendMultipartMessage(router_socket, &msg)

	}

}