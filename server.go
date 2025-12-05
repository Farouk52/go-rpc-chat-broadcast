package main

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

// Message represents a chat message.
type Message struct {
	From string
	Text string
	Time time.Time
}

// RegisterArgs used by client to register.
type RegisterArgs struct {
	ClientID string
}

// RegisterReply returns full history on registration.
type RegisterReply struct {
	History []Message
}

// SendArgs when client sends a message.
type SendArgs struct {
	From string
	Text string
}

// PollArgs used by client to long-poll for messages.
type PollArgs struct {
	ClientID string
}

// ChatServer implements RPC methods.
type ChatServer struct {
	mu      sync.Mutex
	history []Message
	clients map[string]chan Message
}

// NewChatServer initializes server.
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]chan Message),
	}
}

// Register adds a client and returns history. Also broadcasts "joined" to others.
func (s *ChatServer) Register(args RegisterArgs, reply *RegisterReply) error {
	if args.ClientID == "" {
		return errors.New("ClientID required")
	}

	s.mu.Lock()
	// avoid duplicate register: if exists, replace channel
	if _, ok := s.clients[args.ClientID]; ok {
		// close old channel to cleanup
		close(s.clients[args.ClientID])
	}
	ch := make(chan Message, 32) // buffered channel
	s.clients[args.ClientID] = ch

	// copy history to return
	histCopy := make([]Message, len(s.history))
	copy(histCopy, s.history)
	s.mu.Unlock()

	// announce join (do not send to the joining client)
	joinMsg := Message{
		From: "system",
		Text: "User " + args.ClientID + " joined",
		Time: time.Now(),
	}
	// append to history and broadcast
	s.appendAndBroadcast(joinMsg, args.ClientID)

	reply.History = histCopy
	return nil
}

// SendMessage receives a message from a client and broadcasts it.
func (s *ChatServer) SendMessage(args SendArgs, ack *bool) error {
	if args.From == "" {
		return errors.New("From required")
	}
	msg := Message{
		From: args.From,
		Text: args.Text,
		Time: time.Now(),
	}
	s.appendAndBroadcast(msg, args.From)
	*ack = true
	return nil
}

// Poll blocks until the next message for the client is available.
func (s *ChatServer) Poll(args PollArgs, out *Message) error {
	s.mu.Lock()
	ch, ok := s.clients[args.ClientID]
	s.mu.Unlock()
	if !ok {
		return errors.New("client not registered")
	}

	// block until a message is available or channel closed
	msg, ok := <-ch
	if !ok {
		return errors.New("client channel closed")
	}
	*out = msg
	return nil
}

// Unregister removes client and closes its channel.
func (s *ChatServer) Unregister(args RegisterArgs, ack *bool) error {
	s.mu.Lock()
	ch, ok := s.clients[args.ClientID]
	if ok {
		delete(s.clients, args.ClientID)
		close(ch)
	}
	s.mu.Unlock()
	*ack = ok
	return nil
}

// appendAndBroadcast appends to history and broadcasts to all except excludeID.
func (s *ChatServer) appendAndBroadcast(msg Message, excludeID string) {
	// append to history
	s.mu.Lock()
	s.history = append(s.history, msg)

	// copy clients to avoid holding lock while sending
	clientsCopy := make(map[string]chan Message, len(s.clients))
	for id, ch := range s.clients {
		clientsCopy[id] = ch
	}
	s.mu.Unlock()

	// broadcast asynchronously per client to avoid blocking
	for id, ch := range clientsCopy {
		if id == excludeID {
			continue // no self-echo
		}
		// send in goroutine so a slow client won't stall others
		go func(c chan Message) {
			// try to send; if buffer full, this will block — fine for reliability
			c <- msg
		}(ch)
	}
}

func main() {
	srv := NewChatServer()
	err := rpc.Register(srv)
	if err != nil {
		log.Fatalf("rpc register failed: %v", err)
	}

	l, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	log.Println("Chat RPC server listening on :12345")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
