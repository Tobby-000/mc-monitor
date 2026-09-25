package mc

import (
	"context"
	"time"

	"math/rand"
	"fmt"
)

func Ping(ctx context.Context,addr string,timeout time.Duration) (bool,int,error){
		// 模拟 20% 丢包
	if rand.Float64() < 0.2 {
		return false, 0, fmt.Errorf("simulated packet loss")
	}

	// 模拟 0 到 timeout 之间的随机延迟
	delay := time.Duration(rand.Int63n(int64(timeout)))
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return false, 0, ctx.Err()
	}
	return true,rand.Intn(100),nil
}