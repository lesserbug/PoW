package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
)

type Message struct {
	Type      string
	Content   []byte
	Signature []byte
	SenderID  int
}

func (pow *PoW) HandleMessage(msg Message) {
	switch msg.Type {
	case "newBlock":
		pow.handleNewBlock(msg)
	}
}

func (pow *PoW) handleNewBlock(msg Message) {
	pow.miningActive = false
	fmt.Println("Node", pow.node.ID, "received a new block message from node", msg.SenderID)

	senderPubKey := getPubKey(msg.SenderID)
	if !ed25519.Verify(senderPubKey, msg.Content, msg.Signature) {
		fmt.Println("Invalid signature from node", msg.SenderID)
		pow.resumeMining()
		return
	}

	var blockContent struct {
		Block Block
		Hash  string
	}

	err := json.Unmarshal(msg.Content, &blockContent)
	if err != nil {
		fmt.Println("Error unmarshalling block data:", err)
		pow.resumeMining()
		return
	}

	// 验证区块的哈希是否正确
	computedHash := getHash(blockContent.Block)
	if computedHash != blockContent.Hash {
		fmt.Println("Block hash does not match the computed hash")
		pow.resumeMining()
		return
	}

	// 验证区块是否可以逻辑上添加到当前链上
	if len(pow.Chain) > 0 {
		lastBlock := pow.Chain[len(pow.Chain)-1]
		// if blockContent.Block.PrevHash != lastBlock.Hash {
		// 	fmt.Println("Previous hash does not match the hash of the last block in the chain")
		// 	pow.resumeMining()
		// 	return
		// }
		// if blockContent.Block.Height != lastBlock.Height+1 {
		// 	fmt.Println("Block height is not valid")
		// 	pow.resumeMining()
		// 	return
		// }
		if blockContent.Block.PrevHash == lastBlock.Hash && blockContent.Block.Height == lastBlock.Height+1 {
			pow.addBlocktoChain(blockContent.Block)
		} else {
			pow.handleFork(blockContent.Block)
		}
	}
	// pow.addBlocktoChain(blockContent.Block)
}

func (pow *PoW) handleFork(newBlock Block) {
	fmt.Println("Node", pow.node.ID, "detected a fork")
	// 1. Find the common ancestor
	// 2. Roll back the chain to the common ancestor
	// 3. Add the new block to the chain
	// 4. Resume mining
}

func (pow *PoW) resumeMining() {
	pow.mutex.Lock()
	pow.miningActive = true
	pow.mutex.Unlock()
	go pow.MineBlock()
}
