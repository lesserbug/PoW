package main

import (
	"math/rand"
	"time"
)

func main() {
	dc := NewDataCollector()

	powInstances := []*PoW{
		// NewPoW(1, "localhost:8001", dc),
		// NewPoW(2, "localhost:8002", dc),
		// NewPoW(3, "localhost:8003", dc),
		// NewPoW(4, "localhost:8004", dc),
		// NewPoW(5, "localhost:8005", dc),
		// NewPoW(6, "localhost:8006", dc),
		// NewPoW(7, "localhost:8007", dc),
		// NewPoW(8, "localhost:8008", dc),
		// NewPoW(9, "localhost:8009", dc),
		// NewPoW(10, "localhost:8010", dc),

		NewPoW(31, "0.0.0.0:8001", dc),
		NewPoW(32, "0.0.0.0:8002", dc),
		NewPoW(33, "0.0.0.0:8003", dc),
		NewPoW(34, "0.0.0.0:8004", dc),
		NewPoW(35, "0.0.0.0:8005", dc),
		NewPoW(36, "0.0.0.0:8006", dc),
		NewPoW(37, "0.0.0.0:8007", dc),
		NewPoW(38, "0.0.0.0:8008", dc),
		NewPoW(39, "0.0.0.0:8009", dc),
		NewPoW(40, "0.0.0.0:8010", dc),
	}

	// for i, p := range powInstances {
	// 	for j, p2 := range powInstances {
	// 		if i != j {
	// 			p.Nodes = append(p.Nodes, p2.node)
	// 		}
	// 	}
	// }

	ip2 := "54.79.212.120:"

	powInstances2 := []*PoW{
		NewPoW(1, ip2+"8001", dc),
		NewPoW(2, ip2+"8002", dc),
		NewPoW(3, ip2+"8003", dc),
		NewPoW(4, ip2+"8004", dc),
		NewPoW(5, ip2+"8005", dc),
		NewPoW(6, ip2+"8006", dc),
		NewPoW(7, ip2+"8007", dc),
		NewPoW(8, ip2+"8008", dc),
		NewPoW(9, ip2+"8009", dc),
		NewPoW(10, ip2+"8010", dc),
	}

	ip3 := "3.107.25.157:"

        powInstances3 := []*PoW{
                NewPoW(11, ip3+"8001", dc),
                NewPoW(12, ip3+"8002", dc),
                NewPoW(13, ip3+"8003", dc),
                NewPoW(14, ip3+"8004", dc),
                NewPoW(15, ip3+"8005", dc),
                NewPoW(16, ip3+"8006", dc),
                NewPoW(17, ip3+"8007", dc),
                NewPoW(18, ip3+"8008", dc),
                NewPoW(19, ip3+"8009", dc),
                NewPoW(20, ip3+"8010", dc),
        }

	ip4 := "54.206.229.3:"

        powInstances4 := []*PoW{
                NewPoW(21, ip4+"8001", dc),
                NewPoW(22, ip4+"8002", dc),
                NewPoW(23, ip4+"8003", dc),
                NewPoW(24, ip4+"8004", dc),
                NewPoW(25, ip4+"8005", dc),
                NewPoW(26, ip4+"8006", dc),
                NewPoW(27, ip4+"8007", dc),
                NewPoW(28, ip4+"8008", dc),
                NewPoW(29, ip4+"8009", dc),
                NewPoW(30, ip4+"8010", dc),
        }

	for i, p := range powInstances {
		for j, p2 := range powInstances {
			if i != j {
				p.Nodes = append(p.Nodes, p2.node)
			}
		}
	}

	for _, p := range powInstances {
		for _, p2 := range powInstances2 {
			p.Nodes = append(p.Nodes, p2.node)
		}
	}

	for _, p := range powInstances {
                for _, p2 := range powInstances3 {
                        p.Nodes = append(p.Nodes, p2.node)
                }
        }

	for _, p := range powInstances {
                for _, p2 := range powInstances4 {
                        p.Nodes = append(p.Nodes, p2.node)
                }
        }

	for _, p := range powInstances {
		go p.Listen()
		go p.adjustParticipation()
		go p.MineBlock()
	}

	time.Sleep(1 * time.Second) // 等待节点启动
	go collectGeneralDataEvery10Seconds(dc, powInstances)

	go exportData(dc)

	select {}
}

func (pow *PoW) adjustParticipation() {
	start := time.Now()
	for {
		elapsed := time.Since(start)
		var active bool

		timeBlock := int(elapsed.Seconds()) / 10 // 每10秒一个时间块

		switch {
		case elapsed < 400*time.Second:
			// 40-70% 的概率活跃
			active = rand.Float32() < (0.40 + rand.Float32()*0.30)

		case elapsed < 800*time.Second: // 约 125 行
			// 基于时间块的剧烈波动
			if timeBlock%2 == 0 { // 偶数时间块
				active = rand.Float32() < 0.9 // 90% 概率活跃
			} else { // 奇数时间块
				active = rand.Float32() < 0.1 // 10% 概率活跃
			}
			// 添加一些随机性
			if rand.Float32() < 0.1 { // 10% 的概率反转状态
				active = !active
			}

		case elapsed < 1100*time.Second:
			active = rand.Float32() < (0.6 + rand.Float32()*0.3)

		default:
			// 10-30% 的概率活跃
			active = rand.Float32() < (0.1 + rand.Float32()*0.2)
		}

		// 更新该节点的活跃状态
		pow.rwmutex.Lock()
		pow.ActiveNodes[pow.node.ID] = active
		pow.rwmutex.Unlock()

		time.Sleep(10 * time.Second) // 每10秒更新一次
	}
}

// func periodicallyExportData(dc *DataCollector) {
// 	exportTicker := time.NewTicker(3 * time.Minute) // 根据需要调整时间间隔
// 	defer exportTicker.Stop()

// 	for range exportTicker.C {
// 		exportData(dc)
// 	}
// }

func collectGeneralDataEvery10Seconds(dc *DataCollector, powInstances []*PoW) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, pow := range powInstances {
			pow.rwmutex.RLock()
			isActive := pow.ActiveNodes[pow.node.ID]
			logLength := len(pow.Chain)
			pow.rwmutex.RUnlock()

			activeCount := 0
			if isActive {
				activeCount = 1
			}

			// Record data for each node
			dc.RecordGeneralData(activeCount, logLength)
		}
	}
}
