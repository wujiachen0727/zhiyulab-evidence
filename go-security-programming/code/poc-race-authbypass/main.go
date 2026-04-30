// Package main demonstrates a TOCTOU (time-of-check-to-time-of-use) race
// in a typical "业务代码" session cache pattern.
//
// PoC E4: goroutine 隔离保证不了"权限校验完的那一刻 == 操作执行的那一刻"。
//
// Run: go run -race main.go
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Session 是简化的会话信息。
type Session struct {
	UserID string
	Role   string // "user" | "admin"
}

// SessionStore 看起来"加了锁"的会话存储——常见的业务实现。
type SessionStore struct {
	mu    sync.RWMutex
	store map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{store: make(map[string]*Session)}
}

func (s *SessionStore) Get(token string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store[token]
}

func (s *SessionStore) Set(token string, sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[token] = sess
}

// --- 漏洞代码 ---

// adminOperation 检查权限 → 做敏感操作。
// 看起来没问题：先 Get 判断 Role 是 admin，再执行。
// 问题：Get 返回的是 *Session 指针——Get 和 Use 之间，
// 另一个 goroutine 可以**原地修改** sess.Role，
// 或者 Set 了一个同 token 的新 session（指针换了）。
//
// 典型业务场景：管理员临时调整某用户角色，
// 或者 session 刷新时"顺便"把 Role 同步过来。
func adminOperation(store *SessionStore, token string, vault *int64) bool {
	sess := store.Get(token) // T0: 读权限快照
	if sess == nil || sess.Role != "admin" {
		return false
	}

	// T1: 模拟"业务处理耗时"——真实场景里这里可能是 RPC、DB 查询、日志写入
	time.Sleep(1 * time.Millisecond)

	// T2: 使用权限。此刻 sess.Role 可能已被改
	// 但我们认为它还是 "admin"（因为 T0 检查过了）
	atomic.AddInt64(vault, 1000)
	return true
}

// attacker 模拟"另一个 goroutine 在 T0-T2 之间修改了会话"。
// 现实威胁模型：会话数据同步任务、角色调整接口、甚至是另一条登录路径
// 在管理员退出时没清理干净。
func attacker(store *SessionStore, token string) {
	// 把同一个 token 的 session 换成普通用户
	// 如果写的人用了同一个指针（见 sneakyAttacker），问题更严重
	store.Set(token, &Session{UserID: "u-1", Role: "user"})
}

// sneakyAttacker 更阴险的场景：不 Set 新对象，而是在原 Session 上原地改。
// 业务代码里非常常见——session 是指针，"顺手改一下字段"。
func sneakyAttacker(sess *Session) {
	sess.Role = "user"
}

func main() {
	store := NewSessionStore()
	token := "tok-alice"
	// alice 原本是 admin
	adminSess := &Session{UserID: "u-alice", Role: "admin"}
	store.Set(token, adminSess)

	var vault int64
	var wg sync.WaitGroup
	var passes int64

	// 启动一组"管理员操作" goroutine
	const rounds = 500
	for i := 0; i < rounds; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if adminOperation(store, token, &vault) {
				atomic.AddInt64(&passes, 1)
			}
		}()

		// 并发启动一个"偷偷改 Role"的 attacker
		if i%3 == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sneakyAttacker(adminSess) // 原地改，不需要 Set
			}()
		}

		// 以及一个"换新 session"的 attacker
		if i%5 == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				attacker(store, token)
			}()
		}
	}

	wg.Wait()

	fmt.Printf("总共 %d 次 adminOperation，通过权限校验 %d 次\n", rounds, passes)
	fmt.Printf("金库余额：%d（每次通过 +1000）\n", vault)
	fmt.Println()
	fmt.Println("=== 关键观察 ===")
	fmt.Println("1. 用 `go run -race main.go` 运行，race detector 会报 DATA RACE")
	fmt.Println("   指向 sneakyAttacker 对 sess.Role 的写 vs adminOperation 的读")
	fmt.Println("2. 即便没有 race detector，通过次数可能 < rounds（取决于调度）")
	fmt.Println("   如果某次 adminOperation 在 T0 看到 admin，T2 被降权，操作仍然执行")
	fmt.Println("3. goroutine 隔离只隔离栈，不隔离堆上的共享对象")
	fmt.Println("4. sync.RWMutex 保护了 map 访问，没保护 map value 指向的对象的字段")
}
