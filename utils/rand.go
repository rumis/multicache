package utils

import (
	"math/rand"
	"sync"
	"time"
)

var safeRandOnce sync.Once
var safeRandInst *safeRand

// safeRand 线程安全的随机数生成器
type safeRand struct {
	m sync.Mutex
}

// SafeRand 获取线程安全的随机数生成器
func SafeRand() *safeRand {
	safeRandOnce.Do(func() {
		safeRandInst = &safeRand{}
	})
	return safeRandInst
}

// Intn 生成随机数 [0-n)
func (r *safeRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.m.Lock()
	defer r.m.Unlock()

	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n)
}
