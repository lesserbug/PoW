package main

import (
	"crypto/ed25519"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
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

	ProposedTime  time.Time
	ConfirmedTime time.Time
}

type DataCollector struct {
	ActiveCounts []int
	LogLengths   []int
	Latencies    []time.Duration
	rwMutex      sync.RWMutex
}

type PoW struct {
	node         Node
	Nodes        []Node
	ActiveNodes  map[int]bool
	Chain        []Block
	mutex        sync.Mutex
	rwmutex      sync.RWMutex
	miningActive bool // 控制是否正在挖矿

	Data struct {
		ActiveCount []int
		LogLength   []int
		Latencies   []time.Duration
	}
}

func NewPoW(nodeID int, addr string) *PoW {
	p := new(PoW)
	p.miningActive = true
	p.node.ID = nodeID
	p.node.Address = addr
	p.Nodes = make([]Node, 0)
	p.ActiveNodes = make(map[int]bool)
	p.node.PublicKey = getPubKey(nodeID)
	p.node.PrivateKey = getPrivKey(nodeID)
	p.Chain = make([]Block, 0)
	p.Chain = append(p.Chain, createGenesisBlock())

	p.Data.ActiveCount = make([]int, 0)
	p.Data.LogLength = make([]int, 0)
	p.Data.Latencies = make([]time.Duration, 0)

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
		DiffNum:   26, // 初始难度值
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

func (dc *DataCollector) RecordData(activeCount, logLength int, latency time.Duration) {
	dc.rwMutex.Lock()
	defer dc.rwMutex.Unlock()
	dc.ActiveCounts = append(dc.ActiveCounts, activeCount)
	dc.LogLengths = append(dc.LogLengths, logLength)
	dc.Latencies = append(dc.Latencies, latency)
}

func NewDataCollector() *DataCollector {
	return &DataCollector{}
}

// 修改每个节点的 collectData 方法，以便它们能够记录并上报数据
func (pow *PoW) collectData(dc *DataCollector) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pow.rwmutex.RLock()
		isActive := pow.ActiveNodes[pow.node.ID]
		pow.rwmutex.RUnlock()

		activeCount := 0
		if isActive {
			activeCount = 1
		}

		// lastBlock := pow.Chain[len(pow.Chain)-k]
		// if lastBlock.Confirmed {
		// 	latency := lastBlock.ConfirmedTime.Sub(lastBlock.ProposedTime)
		// 	logLength := len(pow.Chain)

		// 	// 这里只记录当前节点的数据，需要一个机制来聚合所有节点的数据
		// 	dc.RecordData(activeCount, logLength, latency)
		// }
		var latency time.Duration
		if len(pow.Chain) > k {
			confirmedBlock := pow.Chain[len(pow.Chain)-k]
			if confirmedBlock.Confirmed {
				latency = confirmedBlock.ConfirmedTime.Sub(confirmedBlock.ProposedTime)
			}
		}
		logLength := len(pow.Chain)
		dc.RecordData(activeCount, logLength, latency)
	}
}

// 导出数据到CSV文件
func (dc *DataCollector) ExportData(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for i, activeCount := range dc.ActiveCounts {
		writer.Write([]string{
			strconv.Itoa(activeCount),
			strconv.Itoa(dc.LogLengths[i]),
			dc.Latencies[i].String(),
		})
	}
}
