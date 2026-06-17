package tools

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Snowflake 配置常量
const (
	epoch       int64 = 1735689600000           // 自定义起始时间戳：2025-01-01 00:00:00 UTC
	workerBits  uint8 = 10                      // 工作节点 ID 占位数
	seqBits     uint8 = 12                      // 序列号占位数
	workerMax   int64 = -1 ^ (-1 << workerBits) // 1023
	seqMax      int64 = -1 ^ (-1 << seqBits)    // 4095
	timeShift   uint8 = workerBits + seqBits    // 22
	workerShift uint8 = seqBits                 // 12
)

// Snowflake 雪花算法生成器
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64 // 上次生成 ID 的时间戳（毫秒）
	workerID  int64 // 工作节点 ID（0~1023）
	seq       int64 // 当前毫秒内的序列号（0~4095）
}

// globalSnowflake 全局雪花算法实例
var globalSnowflake *Snowflake

// init 从环境变量读取 WorkerID 初始化雪花算法
func init() {
	workerID := int64(0)
	if v := os.Getenv("SNOWFLAKE_WORKER_ID"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err == nil && id >= 0 && id <= workerMax {
			workerID = id
		}
	}
	globalSnowflake = NewSnowflake(workerID)
}

// NewSnowflake 创建雪花算法实例
func NewSnowflake(workerID int64) *Snowflake {
	if workerID < 0 || workerID > workerMax {
		panic(fmt.Sprintf("snowflake workerID must be between 0 and %d", workerMax))
	}
	return &Snowflake{
		timestamp: 0,
		workerID:  workerID,
		seq:       0,
	}
}

// NextID 生成下一个唯一 ID
//
// 保证单调递增策略：
//   - 同步互斥锁，线程安全
//   - 正常情况：时间戳前进时 seq 归零；同一毫秒内 seq 递增
//   - 时钟回拨 ≤ 1s：不回退时间戳，复用上次时间戳继续递增 seq
//   - 时钟回拨 > 1s：panic 报警（需运维修复时钟后重启）
func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	// 时钟回拨处理
	if now < s.timestamp {
		diff := s.timestamp - now
		if diff > 1000 { // 回拨超过 1 秒 → 硬失败，强制运维介入
			panic(fmt.Sprintf("clock moved backwards by %d ms (workerID=%d)", diff, s.workerID))
		}
		// 小范围回拨（≤1s）：复用上次时间戳，继续推进 seq，保证 ID 单调递增
		now = s.timestamp
	}

	// 同一毫秒内，序列号递增
	if now == s.timestamp {
		s.seq = (s.seq + 1) & seqMax
		// 当前毫秒序列号用完，等待下一毫秒
		if s.seq == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		// 不同毫秒，序列号重置
		s.seq = 0
	}

	s.timestamp = now

	// 拼接 ID：时间戳左移22位 | WorkerID左移12位 | 序列号
	return ((now - epoch) << timeShift) | (s.workerID << workerShift) | s.seq
}

// GenOrderNo 生成唯一订单号（int64，线程安全）
// 使用方式：orderNo := tools.GenOrderNo()
func GenOrderNo() int64 {
	return globalSnowflake.NextID()
}

// GenOrderNoStr 生成唯一订单号字符串
func GenOrderNoStr() string {
	return strconv.FormatInt(GenOrderNo(), 10)
}

// SnowflakeWorkerID 返回当前 WorkerID（用于日志/监控）
func SnowflakeWorkerID() int64 {
	return globalSnowflake.workerID
}
