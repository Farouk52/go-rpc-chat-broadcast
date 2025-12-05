package main

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

// reuse Message struct
type Message struct {
	From string
	Text string
	Time time.Time
}

type RegisterArgs struct{ ClientID string }
type RegisterReply struct{ History []Message }
type SendArgs struct{ From, Text string }
type PollArgs struct{ ClientID string }

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <your-name>")
		return
	}
	name := os.Args[1]

	client, err := rpc.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}
	defer client.Close()

	// Register
	var regReply RegisterReply
	if err := client.Call("ChatServer.Register", RegisterArgs{ClientID: name}, &regReply); err != nil {
		log.Fatalf("register failed: %v", err)
	}

	// print history
	fmt.Println("=== chat history ===")
	for _, m := range regReply.History {
		fmt.Printf("[%s] %s: %s\n", m.Time.Format("15:04:05"), m.From, m.Text)
	}
	fmt.Println("====================")

	// start poller goroutine
	go func() {
		for {
			var incoming Message
			err := client.Call("ChatServer.Poll", PollArgs{ClientID: name}, &incoming)
			if err != nil {
				log.Printf("poll error: %v", err)
				time.Sleep(time.Second)
				continue
			}
			// print incoming message (no self-echo will be delivered from server)
			fmt.Printf("\n[%s] %s: %s\n> ", incoming.Time.Format("15:04:05"), incoming.From, incoming.Text)
		}
	}()

	// user input loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if txt == "" {
			fmt.Print("> ")
			continue
		}
		if txt == "/quit" {
			var ack bool
			_ = client.Call("ChatServer.Unregister", RegisterArgs{ClientID: name}, &ack)
			fmt.Println("Goodbye.")
			return
		}
		var ack bool
		err := client.Call("ChatServer.SendMessage", SendArgs{From: name, Text: txt}, &ack)
		if err != nil {
			log.Printf("send error: %v", err)
		}
		fmt.Print("> ")
	}
}
