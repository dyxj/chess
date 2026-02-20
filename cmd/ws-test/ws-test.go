package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// Testing new ws lib
func main() {
	mainCtx, mainStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer mainStop()

	server := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleWebSocket),
	}

	go func() {
		fmt.Println("Starting WebSocket server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
		fmt.Println("server closed")
	}()

	<-mainCtx.Done()
	fmt.Println("Shutting down server...")
	mainStop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	fmt.Println("Server stopped")
	time.Sleep(10 * time.Second)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	fmt.Println("New WebSocket connection established")

	// Create connection-specific context
	connCtx, connCancel := context.WithCancel(context.Background())
	defer func() {
		connCancel()
		writeClose(conn)
		cErr := conn.Close()
		if cErr != nil {
			log.Println("Failed to close WebSocket connection" + cErr.Error())
		}
		fmt.Println("WebSocket connection closed")
	}()

	// Start read and write goroutines
	msgChan := goRead(conn, connCtx, connCancel)
	writeDone := goWrite(conn, msgChan)

	// Wait for connection to close
	select {
	case <-connCtx.Done():
		fmt.Println("Connection context done, closing connection")
	case <-r.Context().Done():
		fmt.Println("server closing connection due to server shutdown")
		connCancel()
	}

	<-connCtx.Done()
	fmt.Println("game over")
	<-writeDone
	fmt.Println("write over")
}

func goRead(conn net.Conn, ctx context.Context, cancel context.CancelFunc) chan string {
	msgChan := make(chan string, 100)
	go func() {
		defer close(msgChan)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("exiting go read")
				return
			default:
				fmt.Println("waiting for input")
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						fmt.Println("Connection closed, stopping read")
						return
					}
					if wsCloseErr, ok := errors.AsType[wsutil.ClosedError](err); ok {
						fmt.Println("Connection closed, stopping read" + wsCloseErr.Error())
						return
					}
					log.Printf("Error reading message: %v", err)
					return
				}

				msgStr := string(msg)
				fmt.Printf("Reading message: %s (opcode: %v)\n", msgStr, op)

				if msgStr == "gameover" {
					return
				}

				msgChan <- msgStr
			}
		}
	}()
	return msgChan
}

func goWrite(conn net.Conn, msgChan <-chan string) chan struct{} {
	defer fmt.Println("Write goroutine exiting")
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range msgChan {
			fmt.Printf("Writing message: %s\n", msg)
			msgBytes := []byte(msg)
			err := wsutil.WriteServerMessage(conn, ws.OpText, msgBytes)
			if err != nil {
				// Check for connection closed errors
				if errors.Is(err, net.ErrClosed) {
					fmt.Println("Connection closed, stopping write")
					return
				}

				fmt.Printf("Error writing message: %v\n", err)
				return
			}
		}
	}()
	return done
}

// The first closure message is accepted
func writeClose(conn net.Conn) {
	err := wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusNormalClosure, "server is shutting down"))
	if err != nil {
		// Don't log "use of closed network connection" as it's expected during shutdown
		if !errors.Is(err, net.ErrClosed) {
			fmt.Printf("Error writing close message: %v\n", err)
		}
	}

	err2 := wsutil.WriteServerMessage(conn, ws.OpClose, ws.NewCloseFrameBody(ws.StatusInvalidFramePayloadData, "server is shutting down"))
	if err2 != nil {
		// Don't log "use of closed network connection" as it's expected during shutdown
		if !errors.Is(err2, net.ErrClosed) {
			fmt.Printf("Error writing close message: %v\n", err2)
		}
	}
}
