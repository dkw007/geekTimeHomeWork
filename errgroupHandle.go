package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	group, ctx := errgroup.WithContext(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello Go")
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8889",
	}

	// 利用无缓冲chan 模拟单个服务错误退出
	serverOut := make(chan struct{})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		serverOut <- struct{}{} // 阻塞
	})

	// -- 测试 http server 的启动和退出 --

	// g1 启动http server服务
	// g1 退出后, context 将不再阻塞，g2, g3 都会随之退出
	group.Go(func() error {
		return server.ListenAndServe()
	})

	// g2
	// g2 退出时，调用了 shutdown，g1 也会退出
	group.Go(func() error {
		select {
		case <-serverOut:
			fmt.Println("g2 server closed") // 退出会触发 g.cancel, ctx.done 会收到信号
		case <-ctx.Done():
			fmt.Println("g2 errgroup exit")
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		log.Println("g2 shutting down server...")
		return server.Shutdown(timeoutCtx)
	})

	// g3 linux signal 信号的注册和处理
	// g3 捕获到 os 退出信号将会退出
	// g3 退出后, context 将不再阻塞，g2 会随之退出
	// g2 退出时，调用了 shutdown，g1 会退出
	group.Go(func() error {
		quit := make(chan os.Signal, 1)
		// sigint 用户ctrl+c, sigterm程序退出
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			fmt.Println("g3 exit with context")
			return ctx.Err()
		case sig := <-quit:
			fmt.Println("g3 exit with ctrl_c")
			return errors.Errorf("get os exit: %v", sig)
		}
	})

	// 然后 main 函数中的 g.Wait() 退出，所有协程都会退出
	err := group.Wait()
	fmt.Println("output ", err)
	fmt.Println("ctx err ", ctx.Err())
}
