# Cube / Orc

<img src="assets/orc.jpg" alt="orc" style="width:400px;"/>

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

You may need to edit the `.env` file if the application ports are busy.

### Example Output

```text
Starting Orc worker at localhost:8888
Starting Orc manager at localhost:8000
```

### Examples of API requests // TODO: docs

You can make requests to the manager and workers

#### Create new task

Starting container `test-chapter-9.1` with image `timboring/echo-server:latest` on random port. Container health will be checked periodically.

```bash
curl --location 'http://localhost:8000/tasks' \
--header 'Content-Type: application/json' \
--data '{
    "ID": "a7aa1d44-08f6-443e-9378-f5884311019e",
    "State": 2,
    "Task": {
        "State": 1,
        "ID": "bb1d59ef-9fc1-4e4b-a44d-db571eeed203",
        "Name": "test-chapter-9.1",
        "Image": "timboring/echo-server:latest",
        "ExposedPorts": {
            "7777/tcp": {}
        },
        "PortBindings": {
            "7777/tcp": "7777"
        },
        "HealthCheck": "/health"
    }
}'
```

#### Delete task

```bash
curl --location --request DELETE 'http://localhost:8000/tasks/bb1d59ef-9fc1-4e4b-a44d-db571eeed203' \
--data ''
```

#### Check tasks

```bash
curl --location 'http://localhost:8000/tasks'
```

Example output

```json
[
  {
    "ID": "bb1d59ef-9fc1-4e4b-a44d-db571eeed203",
    "ContainerID": "2bddfc71097fae0f0d885284fed9021e569e5de73caf41710dd919c1f47245f0",
    "Name": "test-chapter-9.1",
    "State": 3,
    "Image": "timboring/echo-server:latest",
    "CPU": 0,
    "Memory": 0,
    "Disk": 0,
    "ExposedPorts": {
      "7777/tcp": {}
    },
    "PortBindings": {
      "7777/tcp": "7777"
    },
    "RestartPolicy": "",
    "StartsAt": "2025-04-17T01:02:57.390244064+03:00",
    "FinishedAt": "2025-04-17T01:07:28.350704419+03:00",
    "CreatedAt": "2025-04-17T01:07:26.336715778+03:00",
    "UpdatedAt": "2025-04-17T01:07:28.350704419+03:00",
    "HealthCheck": "/health",
    "RestartCount": 0,
    "HostPorts": {
      "7777/tcp": [
        {
          "HostIp": "0.0.0.0",
          "HostPort": "32789"
        }
      ]
    }
  }
]
```

#### Get Worker Stats

```bash
curl --location 'http://localhost:8888/stats'
```
