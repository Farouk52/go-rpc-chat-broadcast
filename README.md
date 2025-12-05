# Chat System using Go RPC

A simple chat system built with Go using RPC, goroutines, channels, and mutex.

## Files

- `server.go` - RPC Chat server
- `client.go` - RPC Chat client

## Features

- **RPC-based communication** between clients and server
- **Full chat history** - new clients receive all previous messages
- Multiple clients can connect
- Each user gets a unique ID (User 1, User 2, etc.)
- Join/leave notifications
- Messages are broadcast to all other clients (no self-echo)
- Uses goroutines for concurrent connections
- Channels for message updates
- Mutex for thread-safe client list and history

## How to Run

### Start Server
```powershell
go run server.go
```
