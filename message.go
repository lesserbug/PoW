//

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
	// fmt.Println("Node", pow.node.ID, "received a new block message from node", msg.SenderID)

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
		if blockContent.Block.PrevHash == lastBlock.Hash && blockContent.Block.Height == lastBlock.Height+1 {
			pow.addBlocktoChain(blockContent.Block, pow.dc)
		} else {
			pow.handleFork(blockContent.Block)
		}
	}
}

func (pow *PoW) handleFork(newBlock Block) {
	fmt.Println("Node", pow.node.ID, "detected a fork")

	lastBlock := pow.Chain[len(pow.Chain)-1]

	// 如果新区块的高度小于或等于当前链的最后一个区块，我们需要处理分叉
	if newBlock.Height <= lastBlock.Height {
		// 找到共同祖先
		commonAncestorIndex := -1
		for i := len(pow.Chain) - 1; i >= 0; i-- {
			if pow.Chain[i].Height < newBlock.Height {
				commonAncestorIndex = i
				break
			}
		}

		// 如果找不到共同祖先，忽略这个新区块
		if commonAncestorIndex == -1 {
			fmt.Println("Node", pow.node.ID, "could not find common ancestor, ignoring new block")
			pow.resumeMining()
			return
		}

		// 比较新区块链和当前链的节点 ID
		if newBlock.ProposerID < pow.Chain[commonAncestorIndex+1].ProposerID {
			// 新区块链的提议者 ID 更小，接受新区块链
			fmt.Println("Node", pow.node.ID, "accepting new chain with lower proposer ID")
			pow.Chain = pow.Chain[:commonAncestorIndex+1]
			pow.addBlocktoChain(newBlock, pow.dc)
		} else {
			// 保留当前链
			fmt.Println("Node", pow.node.ID, "keeping current chain with lower or equal proposer ID")
		}
	} else {
		// 新区块的高度更高，直接添加到链上
		pow.addBlocktoChain(newBlock, pow.dc)
	}

	pow.resumeMining()
}

func (pow *PoW) resumeMining() {
	pow.mutex.Lock()
	pow.miningActive = true
	pow.mutex.Unlock()
	go pow.MineBlock()
}
