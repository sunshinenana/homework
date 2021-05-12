package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

////基于errgroup实现一个http server的启动和关闭 ，以及linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
func main() {

	g, ctx := errgroup.WithContext(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("ping", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("pong"))
	})

	serverOut := make(chan struct{})
	mux.HandleFunc("shutdown", func(writer http.ResponseWriter, request *http.Request) {
		serverOut <- struct{}{}
	})

	server := http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           mux,
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
	}

	g.Go(func() error {
		return server.ListenAndServe()
	})

	g.Go(func() error {
		select {
			case <- ctx.Done():
				log.Println("errgroup exit")
			case <- serverOut:
				log.Println("server will out")
		}
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		log.Println("shutting down server")
		return server.Shutdown(timeoutCtx)
	})

	g.Go(func() error {
		quit := make(chan os.Signal, 0)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		select {
			case <- ctx.Done():
				log.Println("go exit")
				return ctx.Err()
			case <- quit:
				log.Println("app stop")
				return os.Exit(1)
		}
	})
	fmt.Printf("errgroup exiting: %+v\n", g.Wait())


}
