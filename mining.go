package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"time"
)

const k = 10

func (pow *PoW) MineBlock() {
	fmt.Println("Node", pow.node.ID, "is mining a new block...")

	newBlock := Block{
		PrevHash:  pow.Chain[len(pow.Chain)-1].Hash,
		Timestamp: time.Now().String(),
		Message:   "Block Message: block_" + strconv.Itoa(pow.node.ID) + time.Now().String(),
		Height:    pow.Chain[len(pow.Chain)-1].Height + 1,
		DiffNum:   pow.Chain[len(pow.Chain)-1].DiffNum, // 继承上一个区块的难度值
		RandomNum: rand.Intn(100000),
		Confirmed: false,
	}

	var nonce int = rand.Intn(100000)
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(newBlock.DiffNum)) // 设置目标难度

	for {
		if !pow.miningActive {
			return
		}
		newBlock.RandomNum = nonce
		// fmt.Println("Node", pow.node.ID, "is trying nonce", nonce)
		newBlock.Hash = getHash(newBlock)
		hashBytes, _ := hex.DecodeString(newBlock.Hash)
		hashInt.SetBytes(hashBytes)
		if hashInt.Cmp(target) == -1 {
			fmt.Println("Node", pow.node.ID, "mined a new block")
			pow.miningActive = false
			pow.broadcastNewBlock(newBlock)
			pow.addBlocktoChain(newBlock)
			break
		}
		nonce++
	}
}

func (pow *PoW) broadcastNewBlock(newBlock Block) {
	content, err := json.Marshal(struct {
		Block Block
		Hash  string
	}{
		Block: newBlock,
		Hash:  newBlock.Hash,
	})

	if err != nil {
		fmt.Println("Error marshalling VSS share and commit data:", err)
		return
	}

	signature := ed25519.Sign(pow.node.PrivateKey, content)

	msg := Message{
		Type:      "newBlock",
		Content:   content,
		Signature: signature,
		SenderID:  pow.node.ID,
	}

	pow.BroadcastMessage(msg)
}

func (pow *PoW) addBlocktoChain(newBlock Block) {
	lastBlock := pow.Chain[len(pow.Chain)-1]
	if newBlock.PrevHash == lastBlock.Hash && newBlock.Height == lastBlock.Height+1 {
		pow.Chain = append(pow.Chain, newBlock)
		if len(pow.Chain) > k {
			pow.Chain[len(pow.Chain)-k-1].Confirmed = true
			fmt.Println("Block at height", pow.Chain[len(pow.Chain)-k-1].Height, "is confirmed by node", pow.node.ID)
		}
		fmt.Println("Node", pow.node.ID, "added a new block to the chain with height", newBlock.Height)

		// pow.mutex.Lock()
		// pow.miningActive = true
		// pow.mutex.Unlock()
		// time.Sleep(1 * time.Second)
		// go pow.MineBlock()
		time.Sleep(1 * time.Second)
		pow.resumeMining()
	}
}

func (pow *PoW) adjustDifficulty() {
	if len(pow.Chain) < 2 {
		return
	}

	lastBlock := pow.Chain[len(pow.Chain)-1]
	penultimateBlock := pow.Chain[len(pow.Chain)-2]

	lastTime, _ := time.Parse(time.RFC3339, lastBlock.Timestamp)
	penultimateTime, _ := time.Parse(time.RFC3339, penultimateBlock.Timestamp)
	interval := lastTime.Sub(penultimateTime)

	targetInterval := 1 * time.Second

	if interval < targetInterval {
		lastBlock.DiffNum++
		fmt.Println("Node", pow.node.ID, "increased the difficulty to", lastBlock.DiffNum)
	} else if lastBlock.DiffNum > 1 {
		lastBlock.DiffNum--
	}
}

// getHash 函数用于计算区块的哈希值
func getHash(block Block) string {
	record := strconv.Itoa(block.Height) + block.PrevHash + block.Message + strconv.Itoa(block.RandomNum) + block.Timestamp
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
