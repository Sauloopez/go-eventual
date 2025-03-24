# go-eventual

A Go service that provides event execution with SQLite and RabbitMQ. The service consumes events from a queue, stores them in a SQLite database, and produces events to appropriate queues. It uses a cron job that runs every minute to auto-execute scheduled events. This allows efficient, scalable and extensible delayed event execution.

## Overview

`go-eventual` is designed to:
- Consume event messages from RabbitMQ
- Store events in a SQLite database with scheduling information
- Execute events at the scheduled time by publishing them to RabbitMQ
- Support different scheduling patterns (specific date, day of week, time of day)

## Features

- **Event Consumption**: Listens to a RabbitMQ queue for new event messages
- **Event Storage**: Stores events in SQLite with scheduling metadata
- **Event Publishing**: Publishes events to RabbitMQ when their scheduled time arrives
- **Scheduled Execution**: Runs a cron job every minute to check for events due for execution
- **Flexible Scheduling**: Supports scheduling by specific date, time of day, and day of week

## Installation

### Prerequisites

- Go 1.23.7 or higher
- RabbitMQ server
- SQLite

### Setup

1. Clone this repository.

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file with the `.env.example` file as a template.

4. Build the application:
   ```bash
   go build -o eventual
   ```

## Usage

### Running the Service

```bash
./eventual
```

This will:
1. Connect to the configured RabbitMQ server
2. Initialize the SQLite database
3. Start consuming events from the specified queue
4. Start the cron job to execute scheduled events

### Event Format

Events should be sent to RabbitMQ in JSON format with the following structure:

```json
{
  "Message": "Event content to be published",
  "ExpectedAt": "2025-05-01",
  "Exchange": "target_exchange",
  "DaySchedule": 1,
  "ExpectedClock": "15:30",
  "Times": 1
}
```

Where:
- `Message`: The content that will be published when the event is executed (required)
- `ExpectedAt`: The specific date for event execution (format: YYYY-MM-DD) (or ExpectedClock)
- `Exchange`: The RabbitMQ exchange to publish to (required)
- `DaySchedule`: Day of week (1-7, where 1 is Monday)
- `ExpectedClock`: Time of day for execution (format: HH:MM)
- `Times`: Number of times to execute the event (default: 1)

## Project Structure

- `/internal/config`: Configuration handling
- `/internal/core`: Core service functionality
- `/internal/cron`: Scheduled task processing
- `/internal/db`: Database models and connection handling
- `/internal/rabbit`: RabbitMQ integration

## Development

### Running Tests

### Database Migrations

Database migrations are handled automatically when the service starts.

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
