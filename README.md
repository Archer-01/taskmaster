# Taskmaster: A Process Control System

Taskmaster is a client/server system that allows its users to monitor and control a number of processes on UNIX-like operating systems.

## Usage

Via docker:
```bash
docker compose up
```

To run locally:

```bash
# Server
go run cmd/server/main.go
```

```bash
# Client
go run cmd/client/main.go
```
