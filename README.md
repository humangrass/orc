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

```bash
curl --location 'http://localhost:8000/tasks' \
--header 'Content-Type: application/json' \
--data '{
    "ID": "266592cd-960d-4091-981c-8c25c44b1018",
    "State": 2,
    "Task": {
        "State": 1,
        "ID": "266592cd-960d-4091-981c-8c25c44b1018",
        "Name": "test-from-api-777",
        "Image": "strm/helloworld-http"
    }
}'
```

#### Delete task

```bash
curl --location --request DELETE 'http://localhost:8000/tasks/266592cd-960d-4091-981c-8c25c44b1018' \
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
        "ID": "266592cd-960d-4091-981c-8c25c44b1018",
        "ContainerID": "fd31597bc7eb16c92cc1b59fc2300d92d58c9795648a71e4d4ac9b5b0ff76b08",
        "Name": "test-from-api-777",
        "State": 3,
        "Image": "strm/helloworld-http",
        "Memory": 0,
        "Disk": 0,
        "ExposedPorts": null,
        "PortBindings": null,
        "RestartPolicy": "",
        "StartsAt": "2025-04-13T12:57:35.5391939+03:00",
        "FinishedAt": "2025-04-13T12:58:16.9053937+03:00",
        "CreatedAt": "2025-04-13T12:58:03.8839293+03:00",
        "UpdatedAt": "2025-04-13T12:58:16.9053937+03:00"
    }
]
```

#### Get Worker Stats

```bash
curl --location 'http://localhost:8888/stats'
```
