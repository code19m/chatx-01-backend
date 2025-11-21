# Email Notifications System

This document describes the email notifications system implemented using Kafka as a message queue.

## Overview

When an admin creates a new user via the API, the system:
1. Publishes a user registration event to Kafka topic `user.registration.email`
2. The notification service consumes these events
3. Sends a welcome email with login credentials to the new user

## Architecture

```
┌─────────────┐     Kafka Event      ┌──────────────────┐     SMTP      ┌────────┐
│   API       │ ─────────────────>   │   Notifications  │ ──────────>   │  Email │
│   Server    │  (user.registration) │   Service        │               │ Server │
└─────────────┘                      └──────────────────┘               └────────┘
```

## Components

### 1. Event Producer (API Server)
- **Location**: `internal/auth/usecase/useruc/usecase.go`
- **Trigger**: When `CreateUser` is called by admin
- **Action**: Publishes `UserRegisteredEvent` to Kafka before hashing the password

### 2. Event Model
- **Location**: `internal/events/user_events.go`
- **Fields**:
  - `email`: User's email address
  - `username`: User's username
  - `password`: Plain text password (only in event, not stored)

### 3. Email Service (Interface-based)
- **Location**: `pkg/email/email.go`
- **Interface**: `Sender` - allows for testing and flexibility
- **Implementation**: `Client` - SMTP client using stdlib `net/smtp`
- **Features**:
  - HTML email templates
  - Welcome email with credentials and security warning

### 4. Notification Use Case
- **Location**: `internal/notifications/usecase/usecase.go`
- **Responsibility**: Business logic for sending welcome emails
- **Dependencies**: Depends on `email.Sender` interface

### 5. Notification Handler (Controller Layer)
- **Location**: `internal/notifications/handler.go`
- **Responsibility**: Parses Kafka events and forwards to use case
- **Thin layer**: No business logic, just event parsing

### 6. Configuration Management
- **Location**: `internal/config/config.go`
- **SMTP Config**: Centralized configuration for all SMTP settings
- **Kafka Config**: Centralized Kafka broker and authentication settings

## Configuration

### Environment Variables

#### API Server (Main Application)
```bash
# Kafka Producer Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_SASL_USERNAME=
KAFKA_SASL_PASSWORD=
```

#### Notification Service
```bash
# SMTP Configuration (Required)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@chatx.code19m.uz

# Kafka Consumer Configuration
KAFKA_BROKERS=localhost:9092
KAFKA_SASL_USERNAME=
KAFKA_SASL_PASSWORD=
KAFKA_GROUP_ID=chatx-notifications
```

## Running the Services

### 1. Start Kafka (if not running)
```bash
# Using Docker Compose
docker-compose up -d kafka zookeeper
```

### 2. Create Kafka Topic
```bash
kafka-topics --create \
  --bootstrap-server localhost:9092 \
  --topic user.registration.email \
  --partitions 1 \
  --replication-factor 1
```

### 3. Start API Server
```bash
go run cmd/main.go http
```

### 4. Start Notification Service
```bash
go run cmd/main.go notifications
```

## Email Template

The welcome email includes:
- **Subject**: "Welcome to ChatX - Your Account Credentials"
- **Content**:
  - Welcome message
  - Username and password
  - Login URL (https://chatx.code19m.uz)
  - **Security warning** to change password immediately after first login
  - Professional HTML formatting with branding

## Testing

### 1. Create a User via API
```bash
curl -X POST http://localhost:9900/auth/users \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "username": "newuser",
    "password": "temppassword123"
  }'
```

### 2. Check Notification Service Logs
You should see logs like:
```
INFO processing user registration event email=newuser@example.com username=newuser
INFO welcome email sent successfully email=newuser@example.com username=newuser
```

### 3. Check User's Email Inbox
The user should receive a welcome email with their credentials.

## Error Handling

- **Kafka Connection Failed**: Service will log error and exit
- **SMTP Connection Failed**: Event consumption fails, Kafka will log error
- **Email Send Failed**: Error is logged, no retry (event is marked as consumed)
- **Invalid Event Data**: Error is logged, event is skipped

## Security Considerations

1. **Password in Transit**:
   - Password is sent via Kafka in plain text
   - Ensure Kafka uses encryption (TLS) in production
   - Consider using Kafka SASL authentication

2. **Email Security**:
   - Email contains plain password
   - Users are warned to change password immediately
   - Consider using temporary passwords with expiration

3. **SMTP Credentials**:
   - Store SMTP password in secure environment variables
   - Use app-specific passwords (e.g., Gmail App Passwords)
   - Never commit credentials to version control

## Production Recommendations

1. **Kafka**:
   - Use SSL/TLS encryption
   - Enable SASL authentication
   - Set up proper replication factor

2. **SMTP**:
   - Use dedicated email service (SendGrid, AWS SES, etc.)
   - Implement rate limiting
   - Monitor email delivery status

3. **Monitoring**:
   - Add metrics for email send success/failure
   - Monitor Kafka consumer lag
   - Set up alerts for failed deliveries

4. **High Availability**:
   - Run multiple instances of notification service
   - Use consumer groups for load balancing
   - Implement health checks

## Deployment

### Using Docker

#### Build Image
```bash
docker build -t chatx -f Dockerfile .
```

#### Run Containers
```bash
# API Server
docker run -d \
  --name chatx-api \
  -p 9900:9900 \
  -e KAFKA_BROKERS=kafka:9092 \
  chatx http

# Notification Service (same image, different command)
docker run -d \
  --name chatx-notifications \
  -e KAFKA_BROKERS=kafka:9092 \
  -e SMTP_HOST=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-password \
  chatx notifications
```

## Troubleshooting

### Emails Not Being Sent

1. **Check Kafka Connection**:
   ```bash
   kafka-console-consumer \
     --bootstrap-server localhost:9092 \
     --topic user.registration.email \
     --from-beginning
   ```

2. **Verify SMTP Settings**:
   - Test SMTP connection manually
   - Check firewall rules (port 587/465)
   - Verify credentials

3. **Check Service Logs**:
   - API server logs for event publishing
   - Notification service logs for consumption

### Events Not Being Consumed

1. **Check Consumer Group Status**:
   ```bash
   kafka-consumer-groups \
     --bootstrap-server localhost:9092 \
     --describe \
     --group chatx-notifications
   ```

2. **Reset Consumer Offset** (if needed):
   ```bash
   kafka-consumer-groups \
     --bootstrap-server localhost:9092 \
     --group chatx-notifications \
     --topic user.registration.email \
     --reset-offsets \
     --to-earliest \
     --execute
   ```

## Future Enhancements

1. **Email Templates**:
   - Support multiple email templates
   - Add password reset emails
   - Account verification emails

2. **Email Queue**:
   - Add retry mechanism with exponential backoff
   - Implement dead letter queue for failed emails

3. **Email Tracking**:
   - Track email delivery status
   - Monitor open rates and click-through rates

4. **Multi-channel Notifications**:
   - SMS notifications
   - Push notifications
   - In-app notifications
