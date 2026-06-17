package tools

import (
	"fmt"
	"sync/atomic"
	"time"
)

// 递增序列，用于订单号生成
var seq int64

// GenOrderNo 生成订单号：时间戳(13位毫秒) + 序列(6位) = 19位数字
func GenOrderNo() string {
	s := atomic.AddInt64(&seq, 1) % 1000000
	ts := time.Now().UnixMilli()
	return fmt.Sprintf("%d%06d", ts, s)
}
