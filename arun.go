package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	_ "runtime/pprof"
	"syscall"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

func sendHTTPRequest() error {
	// Send HTTP request to the central server's /NewSession endpoint
	resp, err := http.Get("http://localhost:8088/v1/NewSession")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func connectToWebSocket() (*websocket.Conn, error) {
	// Connect to the central server WebSocket endpoint
	serverURL := "ws://localhost:8088/v1/NewSession"
	headers := make(http.Header)
	headers.Add("Connection", "Upgrade")
	headers.Add("Upgrade", "websocket")

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, headers)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func handleCommands(ws *websocket.Conn) {
	defer ws.Close()

	// Create a new PTY
	ptmx, tty, err := pty.Open()
	if err != nil {
		log.Println("Failed to open PTY:", err)
		return
	}
	defer ptmx.Close()
	defer tty.Close()

	// Start a new interactive shell in the PTY
	cmd := exec.Command("bash", "-i")
	cmd.Stdout = tty
	cmd.Stdin = tty
	cmd.Stderr = tty

	if err := cmd.Start(); err != nil {
		log.Println("Failed to start shell:", err)
		return
	}
    fmt.Println("oepened the shell")
	// Goroutine to capture output from the PTY and send it to the WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				log.Println("Failed to read from PTY:", err)
				return
			}
			if n > 0 {
				output := string(buf[:n])
				// Send output to WebSocket as a text message
				err = ws.WriteMessage(websocket.TextMessage, []byte(output))
				if err != nil {
					log.Println("Failed to write to WebSocket:", err)
					return
				}
			}
		}
	}()

	// Read commands from the WebSocket and write them to the PTY
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			break
		}
		if len(msg) > 0 {
			// Check for Ctrl+C (ASCII 3 or ETX)
			if string(msg) == "\x03" {
				if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
					log.Println("Failed to send SIGINT:", err)
				}
			} else {
				_, err = ptmx.Write(append(msg, '\n')) // Add newline to simulate Enter key
				if err != nil {
					log.Println("Failed to write to PTY:", err)
					break
				}
			}
		}
	}

	// Wait for the shell process to exit
	if err := cmd.Wait(); err != nil {
		log.Println("Shell process exited with error:", err)
	}
}
var addr = flag.String("addr", "localhost:8088", "http service address")
func main() {

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/v1/NewSession"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//defer c.Close()
	handleCommands(c)

	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// // First, send an HTTP request to the central server's /NewSession endpoint
	// err = sendHTTPRequest()
	// if err != nil {
	// 	log.Fatal("Failed to send HTTP request:", err)
	// }

	// // Connect to the central server WebSocket on the /NewSession endpoint
	// ws, err := connectToWebSocket()
	// if err != nil {
	// 	log.Fatal("Failed to connect to WebSocket server:", err)
	// }

	// // Handle commands received via WebSocket
	// handleCommands(ws)


}
