package main

import (
	"time"
)

func main() {

	powInstances := []*PoW{
		NewPoW(1, "localhost:8001"),
		NewPoW(2, "localhost:8002"),
		NewPoW(3, "localhost:8003"),
	}

	for i, p := range powInstances {
		for j, p2 := range powInstances {
			if i != j {
				p.Nodes = append(p.Nodes, p2.node)
			}
		}
	}

	for _, p := range powInstances {
		go p.Listen()
	}

	time.Sleep(1 * time.Second)

	for _, p := range powInstances {
		go p.MineBlock()
	}

	select {}
}
