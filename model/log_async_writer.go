package model

import (
	"os"
	"sync"

	"github.com/QuantumNous/new-api/common"

	"github.com/bytedance/gopkg/util/gopool"
)

// 异步写入消耗日志：把 LOG_DB.Create 从请求热路径里搬到后台 goroutine，
// 降 P99 延迟。默认关闭（同步写入），由 LOG_CONSUME_ASYNC=true 开启。
//
// 语义：
//   - 启动时 buffer 未满 → 请求协程入队即返回（纳秒级）
//   - buffer 已满 → 回退到同步写（避免丢日志，代价是该请求个别变慢）
//   - 进程退出前可调 StopConsumeLogWriter 等待 drain；默认没注册 graceful
//     shutdown，因为 gin 还没做统一退出钩子——可接受退出时 drop 少量日志。

const consumeLogBufferSize = 1024

var (
	consumeLogChan chan *Log
	consumeLogOnce sync.Once
	consumeLogWG   sync.WaitGroup
)

// StartConsumeLogWriter 启动异步写入协程。幂等安全。
// LOG_CONSUME_ASYNC != "true" 时不启动，enqueueConsumeLog 会自动走同步路径。
func StartConsumeLogWriter() {
	consumeLogOnce.Do(func() {
		if os.Getenv("LOG_CONSUME_ASYNC") != "true" {
			return
		}
		consumeLogChan = make(chan *Log, consumeLogBufferSize)
		consumeLogWG.Add(1)
		gopool.Go(func() {
			defer consumeLogWG.Done()
			for log := range consumeLogChan {
				if err := LOG_DB.Create(log).Error; err != nil {
					common.SysError("async consume log write failed: " + err.Error())
				}
			}
		})
		common.SysLog("consume log async writer started (buffer=" +
			itoa(consumeLogBufferSize) + ")")
	})
}

// StopConsumeLogWriter 关闭 channel 并等待 worker 消费完毕。
// 目前未接 graceful shutdown，仅单测用。
func StopConsumeLogWriter() {
	if consumeLogChan == nil {
		return
	}
	close(consumeLogChan)
	consumeLogWG.Wait()
	consumeLogChan = nil
	consumeLogOnce = sync.Once{} // allow re-Start in tests
}

// enqueueConsumeLog 尝试异步入队；未启用 / buffer 满时回退同步写。
// 返回值仅用于测试断言，生产路径忽略。true=异步入队成功，false=已同步落库。
func enqueueConsumeLog(log *Log) bool {
	if consumeLogChan == nil {
		// 异步未启用 → 同步写
		if err := LOG_DB.Create(log).Error; err != nil {
			common.SysError("sync consume log write failed: " + err.Error())
		}
		return false
	}
	select {
	case consumeLogChan <- log:
		return true
	default:
		// buffer 满，回退同步写避免丢日志
		common.SysError("consume log buffer full, falling back to sync write")
		if err := LOG_DB.Create(log).Error; err != nil {
			common.SysError("sync consume log write failed: " + err.Error())
		}
		return false
	}
}

// itoa 本地化避免 import "strconv" for a single use
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
