# Cube (Orc)

A DIY orchestrator for educational purposes based on the book [Build an Orchestrator in Go (From Scratch)](https://www.manning.com/books/build-an-orchestrator-in-go-from-scratch)

Rewritten for Golang v1.24

# Installation

```bash
git clone git@github.com:humangrass/orc.git
cd orc
go mod download
```

# Usage

### Running the orchestrator

```bash
go run cmd/orc/main.go
```

### Example Output

```text
task: {c9532654-c86e-4372-ac27-7b7a4aa965f1 Task-1 0 Image-1 1024 1 map[] map[]  <nil> <nil> 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401}
task event: {044944f0-c172-4159-8880-8c054f8dbb38 0 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401 {c9532654-c86e-4372-ac27-7b7a4aa965f1 Task-1 0 Image-1 1024 1 map[] map[]  <nil> <nil> 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401} 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401 2025-04-12 15:48:49.7228859 +0300 +03 m=+0.002566401}
worker: {worker-1 {<nil> <nil> 0} map[] 0}
Collecting stats
Starting or stopping a task
Starting a task
Stopping a task
manager: {{<nil> <nil> 0} map[] map[] [worker-1] map[] map[]}
Select Worker...
Update Tasks...
Sending work to workers
node: {Node-1 192.168.1.1 4 1024 0 25 0 worker 0}
create a test container
{"status":"Pulling from library/postgres","id":"13"}
{"status":"Digest: sha256:d714ce760cbf3572aa9f72d9f305f27de79b9e6bdbf81613cd4859df5408831e"}
{"status":"Status: Image is up to date for postgres:13"}
Container a476c4b6913860216887b1c4dc136919f6eb1218edb14c6c1112b62038a10c44 is running with config {test-container-1 false false false map[] [] postgres:13 0 0 0 [POSTGRES_USER=cube POSTGRES_PASSWORD=secret] }
stopping container a476c4b6913860216887b1c4dc136919f6eb1218edb14c6c1112b62038a10c44

2025/04/12 15:48:56 Attempting to stop container a476c4b6913860216887b1c4dc136919f6eb1218edb14c6c1112b62038a10c44
Container a476c4b6913860216887b1c4dc136919f6eb1218edb14c6c1112b62038a10c44 is stopped and removed
```
