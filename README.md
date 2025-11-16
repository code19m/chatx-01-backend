# ChatX Backend

A real-time chat application backend built with Go, following Clean Architecture and principles.

## Prerequisites

- Go 1.25 or higher
- PostgreSQL 16+
- MinIO (for file storage)

## Configuration

Example of `.env` file given on `.env.example` file
Copy the file and rename it to `.env`

## Building

Build the application:

```bash
go build -o chatx ./cmd
```

## Usage

The application provides multiple commands through a single binary:

### Start HTTP Server

```bash
./chatx http
```

Or during development:

```bash
go run ./cmd http
```

### Create Super User

Create an admin user interactively:

```bash
./chatx createsuperuser
```

Or during development:

```bash
go run ./cmd createsuperuser
```

You will be prompted for:

- Email
- Username
- Password
- Password confirmation

### Help

Show available commands:

```bash
./chatx
```

## License

MIT
