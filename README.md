# Notification Service
 
A lightweight event-consumer service for the Medical Scheduling Platform. Part of **AP2 Assignment 3**.
 
The Notification Service has **no gRPC server, no HTTP server, and no database**. Its sole responsibility is to subscribe to domain events published by the Doctor Service and Appointment Service via NATS, deserialize them, and print one structured JSON log line per event to stdout.
 
---
 
## What It Does
 
On startup the service:
1. Reads `NATS_URL` from the environment.
2. Connects to NATS with a retry policy (up to 5 reconnect attempts, 2 s wait between attempts).
3. Subscribes to three subjects: `doctors.created`, `appointments.created`, `appointments.status_updated`.
4. Stays running, consuming and logging messages until stopped.
On each incoming message it:
1. Deserializes the JSON payload.
2. Prints one JSON line to stdout containing `time` (RFC3339), `subject`, and the full `event` object.
3. Returns — no acknowledgement is needed for NATS Core Pub/Sub.
---
 
## Architecture
 
```
NATS Core
  │  doctors.created
  │  appointments.created
  │  appointments.status_updated
  ▼
broker.Client  (nats.go connection with reconnect config)
  │
  ▼
subscriber.Subscriber  (subscribes to all 3 subjects, calls handleMessage)
  │
  ▼
logger.JSONLogger  (marshals EventEnvelope → stdout)
```
 
---
 
## Environment Variables
 
| Variable   | Description       | Example                  |
|------------|-------------------|--------------------------|
| `NATS_URL` | NATS server URL   | `nats://localhost:4222`  |
 
Create a `.env` file in the `notification-service/` directory:
 
```bash
NATS_URL=nats://localhost:4222
```
 
---


---
 
## Error Handling
 
| Situation                           | Behaviour                                                          |
|-------------------------------------|--------------------------------------------------------------------|
| NATS unavailable at startup         | NATS client retries with 2 s wait, up to 5 attempts; if all fail, service exits with non-zero code and a descriptive log message |
| Malformed JSON in a message         | Logs `invalid message subject=<subject> err=<error>` to stdout; message is not silently dropped |
| NATS reconnect during operation     | Handled automatically by the nats.go client (configured reconnect policy) |
 

 
## Project Structure
 
```
notification-service/
├── main.go                        # Entry point: load .env, create App, handle signals
├── .env.example
├── Dockerfile
├── go.mod
└── internal/
    ├── app/                       # Wire-up: broker client, logger, subscriber
    ├── broker/                    # NATS connection with reconnect config
    ├── config/                    # Config struct reading NATS_URL from env
    ├── logger/                    # Logger interface + JSONLogger implementation
    ├── model/                     # EventEnvelope struct (time, subject, event)
    └── subscrider/                # Subscriber: subscribes to 3 subjects, handleMessage
```
 
---

## Start Instruction

Pull [Doctor Service](https://github.com/Aiya594/aitu-ap2-asik1-doctor-service) and [Appointment Service](https://github.com/Aiya594/aitu-ap2-asik1-appointment-service) in the same folder with notification folder. Create `docker-compose.yml` according to  `docker-compose.example.yml` and then:

```bash
docker-compose up -d --build
```


---
 
## Graceful Shutdown
 
The service handles `SIGINT` and `SIGTERM`. On shutdown it:
1. Stops accepting new signals.
2. Closes the NATS connection (in-flight message handlers finish naturally).
3. Exits with code 0.
