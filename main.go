package main

import (
	"math/rand"
	"time"
)

func main() {

	// generateAndSaveKeys(4)
	// generateAndSaveKeys(5)
	// generateAndSaveKeys(6)
	// generateAndSaveKeys(7)
	// generateAndSaveKeys(8)
	// generateAndSaveKeys(9)
	// generateAndSaveKeys(10)

	dc := NewDataCollector()

	powInstances := []*PoW{
		NewPoW(1, "localhost:8001"),
		NewPoW(2, "localhost:8002"),
		NewPoW(3, "localhost:8003"),
		NewPoW(4, "localhost:8004"),
		NewPoW(5, "localhost:8005"),
		// NewPoW(6, "localhost:8006"),
		// NewPoW(7, "localhost:8007"),
		// NewPoW(8, "localhost:8008"),
		// NewPoW(9, "localhost:8009"),
		// NewPoW(10, "localhost:8010"),
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
		go p.adjustParticipation()
		go p.MineBlock()
	}

	time.Sleep(1 * time.Second) // 等待节点启动
	for _, p := range powInstances {
		go p.collectData(dc)
	}

	go periodicallyExportData(dc)

	select {}
}

func (pow *PoW) adjustParticipation() {
	start := time.Now()
	for {
		elapsed := time.Since(start)
		// active := false
		// switch {
		// case elapsed < 600*time.Second:
		// 	active = rand.Float32() < 0.8 // 80% 的几率活跃
		// case elapsed < 1000*time.Second:
		// 	active = rand.Float32() < 0.6 // 60% 的几率活跃
		// case elapsed < 2000*time.Second:
		// 	active = rand.Float32() < 0.5 // 50% 的几率活跃
		// case elapsed < 3000*time.Second:
		// 	active = rand.Float32() < 0.8 // 80% 的几率活跃
		// case elapsed < 3500*time.Second:
		// 	active = rand.Float32() < 0.3 // 30% 的几率活跃
		// default:
		// 	active = rand.Float32() < 0.1 // 10% 的几率活跃
		// }
		// pow.rwmutex.Lock()
		// pow.ActiveNodes[pow.node.ID] = active
		// pow.rwmutex.Unlock()
		// time.Sleep(10 * time.Second) // 每10秒检查一次

		var activeProbability float32

		switch {
		case elapsed < 400*time.Second:
			// 活跃度在50-70之间波动
			activeProbability = 0.5 + rand.Float32()*0.2 // 产生50%到70%的活跃概率
		case elapsed < 800*time.Second:
			// 活跃度在10-90之间剧烈波动
			activeProbability = 0.1 + rand.Float32()*0.8 // 产生10%到90%的活跃概率
		case elapsed < 1200*time.Second:
			// 活跃度迅速下降并保持低稳定
			activeProbability = 0.1 + rand.Float32()*0.1 // 产生10%到20%的活跃概率
		case elapsed < 1600*time.Second:
			// 活跃度在低水平上小幅波动
			activeProbability = 0.05 + rand.Float32()*0.1 // 产生5%到15%的活跃概率
		default:
			// 实验结束后极低活跃度
			activeProbability = 0.01 // 维持1%的活跃概率
		}

		pow.rwmutex.Lock()
		pow.ActiveNodes[pow.node.ID] = rand.Float32() < activeProbability
		pow.rwmutex.Unlock()
		time.Sleep(10 * time.Second) // 每10秒检查一次
	}
}

func periodicallyExportData(dc *DataCollector) {
	exportTicker := time.NewTicker(28 * time.Minute) // 根据需要调整时间间隔
	defer exportTicker.Stop()

	for range exportTicker.C {
		currentTime := time.Now().Format("20060102-150405")
		dc.ExportData("data_" + currentTime + ".csv")
	}
}

// func (pow *PoW) adjustParticipation() {
// 	start := time.Now()
// 	for {
// 		elapsed := time.Since(start)
// 		switch {
// 		case elapsed < 1000*time.Second:
// 			// 前1000秒，活跃度波动不大
// 			setActiveNodes(pow, 6+rand.Intn(3)) //
// 		case elapsed < 2000*time.Second:
// 			// 1000到2000秒，活跃度剧烈变化
// 			setActiveNodes(pow, 5+rand.Intn(6)) // 5到10个节点随机活跃
// 		case elapsed < 3000*time.Second:
// 			// 2000到3000秒，活跃度变化不大
// 			setActiveNodes(pow, 8+rand.Intn(3)) // 所有节点都活跃
// 		case elapsed < 3500*time.Second:
// 			// 3000到3500秒，活跃度剧烈下降
// 			setActiveNodes(pow, 2+rand.Intn(4)) // 2到5个节点随机活跃
// 		default:
// 			// 3500秒后，保持低活跃度
// 			setActiveNodes(pow, 3) // 3个节点活跃
// 		}
// 		time.Sleep(10 * time.Second) // 每10秒检查一次
// 	}
// }

// func setActiveNodes(pow *PoW, count int) {
// 	pow.mutex.Lock()
// 	defer pow.mutex.Unlock()
// 	// 确保所有节点初始状态都是非活跃
// 	for id := 0; id < 10; id++ {
// 		pow.ActiveNodes[id] = false
// 	}
// 	// 随机选择指定数量的节点设为活跃
// 	activeIDs := rand.Perm(10)[:count]
// 	for _, id := range activeIDs {
// 		pow.ActiveNodes[id] = true
// 	}
// }

// func (pow *PoW) collectData() {
// 	go func() {
// 		ticker := time.NewTicker(10 * time.Second) // 每10秒收集一次数据
// 		defer ticker.Stop()
// 		for range ticker.C {
// 			pow.mutex.Lock()
// 			activeCount := 0
// 			for _, active := range pow.ActiveNodes {
// 				if active {
// 					activeCount++
// 				}
// 			}
// 			lastBlock := pow.Chain[len(pow.Chain)-1]
// 			latency := time.Since(lastBlock.ProposedTime)
// 			pow.Data.ActiveCount = append(pow.Data.ActiveCount, activeCount)
// 			pow.Data.LogLength = append(pow.Data.LogLength, len(pow.Chain))
// 			pow.Data.Latencies = append(pow.Data.Latencies, latency)
// 			pow.mutex.Unlock()
// 		}
// 	}()
// }
