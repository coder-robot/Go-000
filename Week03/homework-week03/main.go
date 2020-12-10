package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// unbuffered channel
var done = make(chan int)

func main() {
	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		mux := http.NewServeMux()
		mux.HandleFunc("/close", func(writer http.ResponseWriter, request *http.Request) {
			done <- 1
		})

		// listen port 9001
		serverPtr := NewHttpServer(":9001", mux)
		go func() {
			err := serverPtr.Start()
			if err != nil {
				fmt.Println(err)
			}
		}()

		select {
		case <-done:
			return serverPtr.Stop()
		case <-ctx.Done():
			return errors.New("http server close1! ")
		}
	})

	group.Go(func() error {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-quit:
			return errors.New("signal close! ")
		case <-ctx.Done():
			return errors.New("http server close2! ")
		}
	})

	err := group.Wait()
	fmt.Printf("http server reason: %s\n", err)
}

// http server struct
type HttpServer struct {
	serverPtr *http.Server
	handler   http.Handler
	ctx       context.Context
}

// create new http server ptr
func NewHttpServer(address string, mux http.Handler) *HttpServer {
	httpServerPtr := &HttpServer{ctx: context.Background()}
	httpServerPtr.serverPtr = &http.Server{
		Addr:         address,
		WriteTimeout: 6 * time.Second,
		Handler:      mux,
	}
	return httpServerPtr
}

// http server start
func (httpServerPtr *HttpServer) Start() error {
	fmt.Printf("http sever listen on %s ...\n", httpServerPtr.serverPtr.Addr)
	return httpServerPtr.serverPtr.ListenAndServe()
}

// http server stop
// graceful shutdown
func (httpServerPtr *HttpServer) Stop() error {
	_ = httpServerPtr.serverPtr.Shutdown(httpServerPtr.ctx)
	return fmt.Errorf("http server stopped. ")
}
