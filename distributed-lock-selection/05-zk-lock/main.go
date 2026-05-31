// 实验 E5：ZooKeeper 临时顺序节点 + Watch 前驱节点 实现的分布式锁
//
// 核心机制：
//   1. 所有候选者在 /locks/order/1005/ 下创建 EPHEMERAL_SEQUENTIAL 节点
//      → ZK 给每个节点附加单调递增序号（如 lock-0000000001, lock-0000000002）
//   2. 序号最小的那个 = 当前持锁者
//   3. 其他人 Watch 自己**前一个**节点的删除事件
//      → 持锁者释放（删节点）→ 唯一一个 watcher 被通知 → 抢占公平且无惊群
//   4. 持锁者会话断（崩溃/网络分区）→ ZK 自动删除其 ephemeral 节点 → 锁自动释放
//
// 这一套机制天然就有 fencing：序号本身就是单调递增的 token。
//
// 运行前置：
//   docker run -d --name dl-zk -p 2181:2181 zookeeper:3.9
//
// 运行：
//   go run ./05-zk-lock
package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	rootPath  = "/locks/order/1005"
	connAddr  = "127.0.0.1:2181"
	timeout   = 5 * time.Second
)

func main() {
	t0 := time.Now()

	// 准备根节点
	master, _, err := zk.Connect([]string{connAddr}, timeout)
	if err != nil {
		panic(err)
	}
	ensurePath(master, "/locks")
	ensurePath(master, "/locks/order")
	ensurePath(master, rootPath)
	// 清理旧子节点
	if children, _, err := master.Children(rootPath); err == nil {
		for _, c := range children {
			_ = master.Delete(rootPath+"/"+c, -1)
		}
	}
	master.Close()

	// 三个并发 client 抢锁
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go runClient(i, t0, &wg)
		time.Sleep(200 * time.Millisecond) // 错开启动顺序，便于观察
	}
	wg.Wait()

	fmt.Printf("\n=== 实验结论 ===\n")
	fmt.Printf("- 三个 client 按 EPHEMERAL_SEQUENTIAL 序号依次拿锁\n")
	fmt.Printf("- 持锁者主动释放（或 session 断）→ 后继者立即被通知（Watch 前驱）\n")
	fmt.Printf("- 序号本身就是单调递增 token——天然带 fencing\n")
}

func runClient(id int, t0 time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, _, err := zk.Connect([]string{connAddr}, timeout)
	if err != nil {
		fmt.Printf("[T+%s][C%d] connect err: %v\n", elapsed(t0), id, err)
		return
	}
	defer conn.Close()

	// 1. 创建 EPHEMERAL_SEQUENTIAL 节点
	mySeq, err := conn.Create(rootPath+"/lock-", []byte(fmt.Sprintf("client_%d", id)),
		zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		fmt.Printf("[T+%s][C%d] create err: %v\n", elapsed(t0), id, err)
		return
	}
	myName := mySeq[len(rootPath)+1:]
	fmt.Printf("[T+%s][C%d] 创建顺序节点 %s\n", elapsed(t0), id, myName)

	// 2. 等待轮到自己
	for {
		children, _, err := conn.Children(rootPath)
		if err != nil {
			fmt.Printf("[T+%s][C%d] Children err: %v\n", elapsed(t0), id, err)
			return
		}
		sort.Strings(children)

		// 找到自己在序列中的位置
		idx := -1
		for i, c := range children {
			if c == myName {
				idx = i
				break
			}
		}
		if idx == 0 {
			// 我是最小序号 → 持锁者
			fmt.Printf("[T+%s][C%d] 🟢 我是序号最小（%s）→ 持锁\n", elapsed(t0), id, myName)
			break
		}

		// 不是最小，Watch 我前一个节点的删除
		predecessor := children[idx-1]
		fmt.Printf("[T+%s][C%d] Watch 前驱节点 %s（我是 %s）\n",
			elapsed(t0), id, predecessor, myName)

		exists, _, eventCh, err := conn.ExistsW(rootPath + "/" + predecessor)
		if err != nil {
			fmt.Printf("[T+%s][C%d] ExistsW err: %v\n", elapsed(t0), id, err)
			return
		}
		if !exists {
			// 前驱已经走了，重试一轮
			continue
		}
		// 阻塞等前驱删除事件
		ev := <-eventCh
		if ev.Type == zk.EventNodeDeleted {
			fmt.Printf("[T+%s][C%d] 收到前驱 %s 删除事件，重新检查序列\n",
				elapsed(t0), id, predecessor)
		}
	}

	// 3. 模拟业务工作
	work := time.Duration(500+id*200) * time.Millisecond
	fmt.Printf("[T+%s][C%d] 业务工作 %v\n", elapsed(t0), id, work)
	time.Sleep(work)

	// 4. 释放：删除自己的节点
	if err := conn.Delete(mySeq, -1); err != nil {
		fmt.Printf("[T+%s][C%d] Delete err: %v\n", elapsed(t0), id, err)
		return
	}
	fmt.Printf("[T+%s][C%d] 🔓 已释放锁（删除 %s）\n", elapsed(t0), id, myName)
}

func ensurePath(conn *zk.Conn, path string) {
	exists, _, _ := conn.Exists(path)
	if !exists {
		conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
	}
}

func elapsed(t0 time.Time) string {
	return fmt.Sprintf("%5.2fs", time.Since(t0).Seconds())
}
