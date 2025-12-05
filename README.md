# go-rpc-chat-broadcast

Simple example chat application using Go's `net/rpc` package with a
broadcast-style server. This repository contains a minimal server and
client demonstrating RPC-based message broadcasting between connected
clients.

Files
- `server.go`: RPC server that manages registered clients, broadcasts
	messages, and serves poll requests.
- `client.go`: Simple command-line client that registers, polls for new
	messages, and sends messages to the server.

Quick start

1. Start the server (listen on `127.0.0.1:12345`):

```powershell
go run server.go
```

2. In another terminal, run a client (replace `alice` with your name):

```powershell
go run client.go alice
```

3. Start additional clients in other terminals using different names.
	 Type messages and press Enter to send. Use `/quit` to unregister and
	 exit.

Build binaries

```powershell
go build -o server.exe server.go
go build -o client.exe client.go

# then run
.\server.exe
.\client.exe alice
```

Notes
- The server and client use a simple polling RPC model. Clients call
	`ChatServer.Poll` to receive incoming messages (the client will not
	receive its own messages echoed back by the server).
- This example is intended for learning and demonstration, not for
	production use. Consider using authenticated connections and a more
	robust protocol for real deployments.

License
- MIT-style use (no license file included). Feel free to add a
	`LICENSE` if you want to clarify terms.