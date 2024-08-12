package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"
)

type Node struct {
	ID         int
	Address    string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

type Block struct {
	PrevHash  string
	Hash      string
	Message   string
	Timestamp string
	Height    int
	DiffNum   uint
	RandomNum int
	Confirmed bool
}

type PoW struct {
	node         Node
	Nodes        []Node
	Chain        []Block
	mutex        sync.Mutex
	miningActive bool // 控制是否正在挖矿
}

func NewPoW(nodeID int, addr string) *PoW {
	p := new(PoW)
	p.miningActive = true
	p.node.ID = nodeID
	p.node.Address = addr
	p.Nodes = make([]Node, 0)
	p.node.PublicKey = getPubKey(nodeID)
	p.node.PrivateKey = getPrivKey(nodeID)
	p.Chain = make([]Block, 0)
	p.Chain = append(p.Chain, createGenesisBlock())

	return p
}

// 创建创世区块
func createGenesisBlock() Block {
	genesisBlock := Block{
		PrevHash:  "",
		Hash:      "11111111111111111111111111111111",
		Message:   "Genesis Block",
		Timestamp: time.Now().String(),
		Height:    0,
		DiffNum:   30, // 初始难度值
		RandomNum: 0,
		Confirmed: true,
	}
	// genesisBlock.Hash = getHash(genesisBlock) // 计算创世区块的哈希
	return genesisBlock
}

func (p *PoW) Listen() {
	listener, err := net.Listen("tcp", p.node.Address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		log.Panic(err)
	}
	defer listener.Close()

	fmt.Println("Node", p.node.ID, "is listening on", p.node.Address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go p.HandleConnection(conn)
	}
}

func (p *PoW) HandleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Error decoding message:", err)
		return
	}

	p.HandleMessage(msg)
}

func (p *PoW) SendMessage(address string, msg Message) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(msg)
	if err != nil {
		fmt.Println("Error encoding and sending message:", err)
	}
}

func (pow *PoW) BroadcastMessage(msg Message) {
	for _, node := range pow.Nodes {
		if node.ID != pow.node.ID { // 避免给自己发送消息
			pow.SendMessage(node.Address, msg)
		}
	}
}

func getPubKey(nodeID int) ed25519.PublicKey {
	keyPath := fmt.Sprintf("Keys/Node%d/Node%d_ED25519_PUB", nodeID, nodeID)
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Panic("Failed to read public key:", err)
	}
	return key
}

func getPrivKey(nodeID int) ed25519.PrivateKey {
	keyPath := fmt.Sprintf("Keys/Node%d/Node%d_ED25519_PRIV", nodeID, nodeID)
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Panic("Failed to read private key:", err)
	}
	return key
}
