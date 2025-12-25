package pkg

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const gracefulTimeout = 10 * time.Second

// Graceful 等待 docker stop 触发的终止信号（SIGTERM/SIGINT），在收到信号后顺序执行传入任务。
// 每个任务会共享一个带超时的上下文（默认 10s，与 docker stop 默认宽限期一致）。
func Graceful(tasks ...func(context.Context) error) {
	go graceful(tasks...)
}

func graceful(tasks ...func(context.Context) error) {
	defer os.Exit(0)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)

	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	for _, task := range tasks {
		if task == nil {
			continue
		}
		if err := task(ctx); err != nil {
			fmt.Printf("failed to execute graceful task: %v\n", err)
			return
		}
	}
}
