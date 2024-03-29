package main

import (
	"fmt"
	"math/rand"
	"os"
	. "rtclbedit/curp"
	. "rtclbedit/shared"
	"strconv"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	args := os.Args
	if len(os.Args) != 4 {
		fmt.Println("[usage]: " + args[0] + " <identifier> <topo_file> <num_nodes>")
		os.Exit(1)
	}
	identifier := args[1]
	topo_file := args[2]
	num_nodes, err := strconv.Atoi(args[3])

	if err != nil {
		fmt.Println("Must have an integer node count!")
		os.Exit(1)
	}

	node_map := Parse(topo_file, num_nodes)
	master_node, peer_map, witness_map := ParseByRole(node_map)
	curr_node := node_map[identifier]

	if master_node == nil {
		fmt.Println("Invalid topo file (needs at least one master node)!")
		os.Exit(1)
	}

	if curr_node == nil {
		fmt.Println("Invalid topology (identifier not found in topo file)!")
		os.Exit(1)
	}

	InitRPC(identifier, node_map)

	DPrintf("%s: pre-switch", identifier)
	channel := make(chan ExecuteMsg)
	switch curr_node.Role {
	case ROLE_MASTER:
		c := InitCurp(identifier, peer_map, witness_map, ROLE_MASTER, channel)
		go c.CurpLifetime() // busy-wait forever
	case ROLE_BACKUP:
		c := InitCurp(identifier, peer_map, witness_map, ROLE_BACKUP, channel)
		go c.CurpLifetime() // busy-wait forever
	case ROLE_WITNESS:
		w := InitWitness(identifier, master_node)
		go w.WitnessLifetime() // busy-wait forever
	default:
		panic("Unknown Role! Exiting...")
	}
	DPrintf("%s: post-switch", identifier)

	for {
		exe_msg := <-channel
		DPrintf("Server %s app channel %v+", identifier, exe_msg)
	}
}
