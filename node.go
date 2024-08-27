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

	ProposerID int
}

type DataCollector struct {
	rwMutex     sync.RWMutex
	GeneralData []GeneralDataEntry
}

type GeneralDataEntry struct {
	Timestamp    string
	ActiveCount  int
	LogLength    int
	BlockHeight  int // 这些只在有延迟数据时填写
	ProposedTime string
	Latency      string
}

type PoW struct {
	node         Node
	Nodes        []Node
	ActiveNodes  map[int]bool
	Chain        []Block
	mutex        sync.Mutex
	rwmutex      sync.RWMutex
	miningActive bool // 控制是否正在挖矿

	dc *DataCollector
}

func NewPoW(nodeID int, addr string, dc *DataCollector) *PoW {
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

	p.dc = dc

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
		DiffNum:   25, // 初始难度值
		RandomNum: 0,
		Confirmed: true,
	}
	// genesisBlock.Hash = getHash(genesisBlock) // 计算创世区块的哈希
	return genesisBlock
}

func (p *PoW) Listen() {
	listener, err := net.Listen("tcp", p.node.Address)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
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

//	func (dc *DataCollector) RecordData(activeCount, logLength int) {
//		dc.rwMutex.Lock()
//		defer dc.rwMutex.Unlock()
//		dc.ActiveCounts = append(dc.ActiveCounts, activeCount)
//		dc.LogLengths = append(dc.LogLengths, logLength)
//	}
func (dc *DataCollector) RecordGeneralData(activeCount, logLength int) {
	dc.rwMutex.Lock()
	dc.GeneralData = append(dc.GeneralData, GeneralDataEntry{
		Timestamp:   time.Now().Format(time.RFC3339),
		ActiveCount: activeCount,
		LogLength:   logLength,
	})
	dc.rwMutex.Unlock()
}

func (dc *DataCollector) RecordLatencyData(height int, proposedTime time.Time, latency time.Duration) {
	entry := GeneralDataEntry{
		Timestamp:    time.Now().Format(time.RFC3339),
		BlockHeight:  height,
		ProposedTime: proposedTime.Format(time.RFC3339),
		Latency:      latency.String(),
	}
	dc.rwMutex.Lock()
	dc.GeneralData = append(dc.GeneralData, entry)
	dc.rwMutex.Unlock()
}

func NewDataCollector() *DataCollector {
	return &DataCollector{}
}

// 导出数据到CSV文件
func exportData(dc *DataCollector) {
	ticker := time.NewTicker(20 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		filename := "combined_data_" + time.Now().Format("20060102-150405") + ".csv"
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		writer := csv.NewWriter(file)
		// 写入列标题
		headers := []string{"Timestamp", "ActiveCount", "LogLength", "BlockHeight", "ProposedTime", "Latency"}
		writer.Write(headers)

		for _, data := range dc.GeneralData {
			row := []string{data.Timestamp, strconv.Itoa(data.ActiveCount), strconv.Itoa(data.LogLength), strconv.Itoa(data.BlockHeight), data.ProposedTime, data.Latency}
			writer.Write(row)
		}
		writer.Flush()
		file.Close()

		// 清除已导出数据
		dc.rwMutex.Lock()
		dc.GeneralData = nil
		dc.rwMutex.Unlock()
	}
}
