# ChatX API Contract

Version: 1.1
Base URL: `http://localhost:9900` (configurable via `SERVER_ADDR` env variable)

## Table of Contents

- [Authentication](#authentication)
- [Common Patterns](#common-patterns)
- [Error Responses](#error-responses)
- [Authentication Endpoints](#authentication-endpoints)
- [User Management Endpoints](#user-management-endpoints)
- [Image Management Endpoints](#image-management-endpoints)
- [Chat Endpoints](#chat-endpoints)
- [Message Endpoints](#message-endpoints)
- [Notification Endpoints](#notification-endpoints)
- [WebSocket API](#websocket-api)
- [Data Models](#data-models)
- [Environment Configuration](#environment-configuration)

---

## Authentication

Most endpoints require authentication using JWT Bearer tokens.

**Header Format:**

```bash
Authorization: Bearer <access_token>
```

**Token Types:**

- **Access Token**: Short-lived (15 minutes default), used for API requests
- **Refresh Token**: Long-lived (24 hours default), used to obtain new access tokens

**Roles:**

- `user`: Regular user (default)
- `admin`: Administrator with elevated privileges

---

## Common Patterns

### Pagination

List endpoints support pagination via query parameters:

- `page` (int): Page number (0-indexed, default: 0)
- `limit` (int): Items per page (1-100, default varies by endpoint)

**Response includes:**

```json
{
  "items": [...],
  "total": 150,
  "page": 0,
  "limit": 20
}
```

### Date/Time Format

All timestamps are returned in RFC3339 format:

```
2025-01-15T14:30:00Z
```

### Nullable Fields

Fields that can be `null` are marked with `*` in type definitions and use `omitempty` in JSON responses.

---

## Error Responses

All error responses follow this format:

### Validation Errors (400 Bad Request)

```json
{
  "error": "validation error",
  "fields": {
    "email": "invalid email format",
    "password": "password is required"
  }
}
```

### Authentication Errors (401 Unauthorized)

```json
{
  "error": "unauthorized: invalid token"
}
```

### Authorization Errors (403 Forbidden)

```json
{
  "error": "forbidden: insufficient permissions"
}
```

### Not Found Errors (404 Not Found)

```json
{
  "error": "resource not found"
}
```

### Conflict Errors (409 Conflict)

```json
{
  "error": "resource already exists"
}
```

### Server Errors (500 Internal Server Error)

```json
{
  "error": "internal server error"
}
```

---

## Authentication Endpoints

### POST /auth/login

Login with email and password to receive access and refresh tokens.

**Authentication:** None required

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Validation Rules:**

- `email`: Valid email format required
- `password`: Required, non-empty

**Success Response (200 OK):**

```json
{
  "user_id": 1,
  "username": "johndoe",
  "email": "user@example.com",
  "role": "user",
  "image_path": "path/to/profile.jpg",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Notes:**

- `image_path` can be `null` if user hasn't uploaded a profile image
- `role` will be either `"user"` or `"admin"`

---

### POST /auth/logout

Logout the current user (invalidates tokens on client side).

**Authentication:** Required (Bearer token)

**Request Body:** Empty `{}`

**Success Response (200 OK):** Empty response

---

## User Management Endpoints

### POST /auth/users

Create a new user (admin only).

**Authentication:** Required (Admin role)

**Request Body:**

```json
{
  "email": "newuser@example.com",
  "username": "newuser",
  "password": "password123"
}
```

**Validation Rules:**

- `email`: Valid email format
- `username`: 3-30 characters, alphanumeric and underscore only
- `password`: Required

**Success Response (201 Created):**

```json
{
  "user_id": 42
}
```

---

### GET /auth/users

Get paginated list of all users (admin only).

**Authentication:** Required (Admin role)

**Query Parameters:**

- `page` (int, optional): Page number (default: 0)
- `limit` (int, optional): Items per page (1-100, default: 20)

**Success Response (200 OK):**

```json
{
  "users": [
    {
      "user_id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "role": "user",
      "image_path": "path/to/image.jpg",
      "created_at": "2025-01-15T10:00:00Z"
    }
  ],
  "total": 150,
  "page": 0,
  "limit": 20
}
```

---

### GET /auth/users/{user_id}

Get a specific user's details (admin only).

**Authentication:** Required (Admin role)

**Path Parameters:**

- `user_id` (int): User ID

**Success Response (200 OK):**

```json
{
  "user_id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "role": "user",
  "image_path": "path/to/image.jpg",
  "created_at": "2025-01-15T10:00:00Z"
}
```

---

### DELETE /auth/users/{user_id}

Delete a user (admin only).

**Authentication:** Required (Admin role)

**Path Parameters:**

- `user_id` (int): User ID to delete

**Success Response (200 OK):** Empty response

---

### GET /auth/users/me

Get the authenticated user's profile.

**Authentication:** Required

**Success Response (200 OK):**

```json
{
  "user_id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "role": "user",
  "image_path": "path/to/image.jpg"
}
```

---

### PUT /auth/users/me/password

Change the authenticated user's password.

**Authentication:** Required

**Request Body:**

```json
{
  "old_password": "currentpassword",
  "new_password": "newsecurepassword"
}
```

**Validation Rules:**

- `old_password`: Required
- `new_password`: Required, minimum 8 characters

**Success Response (200 OK):** Empty response

---

### PUT /auth/users/me/image

Update the authenticated user's profile image.

**Authentication:** Required

**Request Body:**

```json
{
  "image_path": "path/to/new/image.jpg"
}
```

**Validation Rules:**

- `image_path`: Required, non-empty string

**Success Response (200 OK):**

```json
{
  "image_path": "path/to/new/image.jpg"
}
```

**Notes:**

- The actual file upload mechanism is separate (likely via MinIO presigned URLs)
- This endpoint updates the reference to an already-uploaded image

---

## Image Management Endpoints

### POST /auth/images/upload

Upload a profile image for the authenticated user.

**Authentication:** Required

**Request:** Multipart form data

**Form Fields:**

- `file`: Image file (JPEG or PNG)

**Validation Rules:**

- File must be present
- Content-Type must be `image/jpeg`, `image/jpg`, or `image/png`
- Maximum file size: 10 MB

**Success Response (200 OK):**

```json
{
  "image_path": "users/1/profile.jpg"
}
```

**Error Responses:**

- `400 Bad Request`: Invalid file format or missing file
- `413 Payload Too Large`: File exceeds 10 MB limit

**Notes:**

- Uploaded images are stored at path: `users/{user_id}/profile.{ext}`
- Each user can have one profile image (uploading a new one replaces the old path)
- After uploading, use `PUT /auth/users/me/image` to update the user's profile with the returned path

**Example cURL:**

```bash
curl -X POST http://localhost:9900/auth/images/upload \
  -H "Authorization: Bearer <your_token>" \
  -F "file=@/path/to/image.jpg"
```

---

### GET /auth/images/{image_path...}

Download an image by its path.

**Authentication:** Not required (public access)

**Path Parameters:**

- `image_path` (string): The full path to the image (e.g., `users/1/profile.jpg`)

**Success Response (200 OK):**

Returns the image file with appropriate headers:

- `Content-Type`: Image MIME type (e.g., `image/jpeg`, `image/png`)
- `Content-Disposition`: `inline; filename="profile.jpg"`

**Error Responses:**

- `404 Not Found`: Image does not exist

**Notes:**

- This endpoint serves images for display in browsers or download
- No authentication required - images are publicly accessible
- The `{image_path...}` wildcard allows for nested paths (e.g., `users/1/profile.jpg`)

**Example Usage:**

```bash
# Download image
curl http://localhost:9900/auth/images/users/1/profile.jpg -o profile.jpg

# Or simply open in browser:
# http://localhost:9900/auth/images/users/1/profile.jpg
```

---

## Chat Endpoints

### GET /chat/chats/dms

Get list of direct message conversations.

**Authentication:** Required

**Query Parameters:**

- `page` (int, optional): Page number (default: 0)
- `limit` (int, optional): Items per page (1-100, default: 20)

**Success Response (200 OK):**

```json
{
  "dms": [
    {
      "chat_id": 1,
      "other_user_id": 2,
      "other_username": "janedoe",
      "other_user_image": "path/to/jane.jpg",
      "last_message_text": "Hey, how are you?",
      "last_message_sent_at": "2025-01-15T14:30:00Z",
      "unread_count": 3
    }
  ],
  "total": 15,
  "page": 0,
  "limit": 20
}
```

**Notes:**

- `other_user_image`, `last_message_text`, and `last_message_sent_at` can be `null`
- `unread_count` shows messages not yet read by the current user

---

### GET /chat/chats/groups

Get list of group conversations.

**Authentication:** Required

**Query Parameters:**

- `page` (int, optional): Page number (default: 0)
- `limit` (int, optional): Items per page (1-100, default: 20)

**Success Response (200 OK):**

```json
{
  "groups": [
    {
      "chat_id": 10,
      "name": "Team Chat",
      "creator_id": 1,
      "participant_count": 5,
      "last_message_text": "Meeting at 3 PM",
      "last_message_sent_at": "2025-01-15T14:30:00Z",
      "unread_count": 2
    }
  ],
  "total": 5,
  "page": 0,
  "limit": 20
}
```

**Notes:**

- `last_message_text` and `last_message_sent_at` can be `null`

---

### GET /chat/chats/{chat_id}

Get detailed information about a specific chat.

**Authentication:** Required

**Path Parameters:**

- `chat_id` (int): Chat ID

**Success Response (200 OK):**

```json
{
  "chat_id": 1,
  "type": "direct",
  "name": "",
  "creator_id": 0,
  "participants": [
    {
      "user_id": 1,
      "username": "johndoe",
      "image_path": "path/to/john.jpg",
      "joined_at": "2025-01-10T10:00:00Z"
    },
    {
      "user_id": 2,
      "username": "janedoe",
      "image_path": null,
      "joined_at": "2025-01-10T10:00:00Z"
    }
  ],
  "created_at": "2025-01-10T10:00:00Z"
}
```

**Notes:**

- `type` is either `"direct"` or `"group"`
- For direct chats: `name` is empty, `creator_id` is 0
- For group chats: `name` contains the group name, `creator_id` shows who created it

---

### POST /chat/chats/dms

Create a new direct message conversation.

**Authentication:** Required

**Request Body:**

```json
{
  "other_user_id": 2
}
```

**Validation Rules:**

- `other_user_id`: Must be > 0

**Success Response (201 Created):**

```json
{
  "chat_id": 15
}
```

**Notes:**

- If a DM already exists between the two users, returns the existing chat_id
- Cannot create DM with yourself (enforced at business logic layer)

---

### POST /chat/chats/groups

Create a new group chat.

**Authentication:** Required

**Request Body:**

```json
{
  "name": "Project Team",
  "participant_ids": [2, 3, 4]
}
```

**Validation Rules:**

- `name`: Required, 1-100 characters
- `participant_ids`: Required, at least one participant

**Success Response (201 Created):**

```json
{
  "chat_id": 20
}
```

**Notes:**

- Creator is automatically added as a participant
- Duplicate user IDs are handled gracefully

---

## Message Endpoints

### GET /chat/chats/{chat_id}/messages

Get paginated messages for a specific chat.

**Authentication:** Required

**Path Parameters:**

- `chat_id` (int): Chat ID

**Query Parameters:**

- `page` (int, optional): Page number (default: 0)
- `limit` (int, optional): Items per page (1-100, default: 50)

**Success Response (200 OK):**

```json
{
  "messages": [
    {
      "message_id": 101,
      "chat_id": 1,
      "sender_id": 2,
      "sender_name": "janedoe",
      "sender_image": "path/to/jane.jpg",
      "content": "Hello there!",
      "sent_at": "2025-01-15T14:30:00Z",
      "edited_at": null
    }
  ],
  "total": 250,
  "page": 0,
  "limit": 50
}
```

**Notes:**

- Messages are ordered by `sent_at` descending (newest first)
- `edited_at` is `null` if message was never edited
- `sender_image` can be `null`
- Deleted messages are not returned in the list

---

### POST /chat/messages

Send a new message.

**Authentication:** Required

**Request Body:**

```json
{
  "chat_id": 1,
  "content": "Hello everyone!"
}
```

**Validation Rules:**

- `chat_id`: Must be > 0
- `content`: Required, 1-5000 characters

**Success Response (201 Created):**

```json
{
  "message_id": 102,
  "sent_at": "2025-01-15T14:35:00Z"
}
```

---

### PUT /chat/messages/{message_id}

Edit an existing message.

**Authentication:** Required

**Path Parameters:**

- `message_id` (int): Message ID

**Request Body:**

```json
{
  "content": "Updated message content"
}
```

**Validation Rules:**

- `content`: Required, 1-5000 characters

**Success Response (200 OK):** Empty response

**Notes:**

- Only the message sender can edit their messages
- Sets `edited_at` timestamp

---

### DELETE /chat/messages/{message_id}

Delete a message.

**Authentication:** Required

**Path Parameters:**

- `message_id` (int): Message ID

**Success Response (200 OK):** Empty response

**Notes:**

- Only the message sender can delete their messages
- Soft delete: message is removed from listings

---

## Notification Endpoints

### GET /chat/notifications/unread

Get total unread message count across all chats.

**Authentication:** Required

**Success Response (200 OK):**

```json
{
  "total_unread_count": 12
}
```

---

### GET /chat/chats/{chat_id}/unread

Get unread message count for a specific chat.

**Authentication:** Required

**Path Parameters:**

- `chat_id` (int): Chat ID

**Success Response (200 OK):**

```json
{
  "chat_id": 1,
  "unread_count": 5
}
```

---

### POST /chat/chats/read

Mark messages as read up to a specific message.

**Authentication:** Required

**Request Body:**

```json
{
  "chat_id": 1,
  "message_id": 150
}
```

**Validation Rules:**

- `chat_id`: Must be > 0
- `message_id`: Must be > 0

**Success Response (200 OK):** Empty response

**Notes:**

- Marks all messages up to and including `message_id` as read
- Updates `last_read_message_id` for the user in this chat

---

### POST /chat/users/online-status

Get online status for multiple users.

**Authentication:** Required

**Request Body:**

```json
{
  "user_ids": [1, 2, 3, 4, 5]
}
```

**Validation Rules:**

- `user_ids`: Required, at least one user, maximum 100 users

**Success Response (200 OK):**

```json
{
  "statuses": [
    {
      "user_id": 1,
      "is_online": true,
      "last_seen": null
    },
    {
      "user_id": 2,
      "is_online": false,
      "last_seen": "2025-01-15T13:00:00Z"
    }
  ]
}
```

**Notes:**

- `last_seen` is `null` if user is currently online
- `last_seen` contains timestamp of last activity when offline
- Online status is now tracked via WebSocket connections (see [WebSocket API](#websocket-api))

---

## WebSocket API

ChatX provides real-time messaging capabilities via WebSocket connections. This allows clients to receive instant notifications for new messages, message edits/deletes, typing indicators, and user presence updates.

### Connection

**Endpoint:** `GET /chat/ws`

**Authentication:** Token via query parameter

**Connection URL:**

```
ws://localhost:9900/chat/ws?token=<access_token>
```

**Notes:**

- Use `wss://` for production environments with TLS
- The `token` query parameter must be a valid JWT access token
- Upon successful connection, the client is automatically subscribed to all chats they participate in
- Connection triggers `presence.online` event to all contacts

---

### Message Format

All WebSocket messages use JSON format with the following envelope structure:

**Server to Client:**

```json
{
  "type": "event_type",
  "payload": { ... }
}
```

**Client to Server:**

```json
{
  "type": "event_type",
  "payload": {
    "chat_id": 1
  }
}
```

---

### Server Events (Server to Client)

#### message.new

Received when a new message is sent in a chat you participate in.

```json
{
  "type": "message.new",
  "payload": {
    "id": 123,
    "chat_id": 1,
    "sender_id": 2,
    "content": "Hello there!",
    "sent_at": "2025-01-15T14:30:00Z"
  }
}
```

---

#### message.edit

Received when a message is edited.

```json
{
  "type": "message.edit",
  "payload": {
    "id": 123,
    "chat_id": 1,
    "sender_id": 2,
    "content": "Updated message content",
    "edited_at": "2025-01-15T14:35:00Z"
  }
}
```

---

#### message.delete

Received when a message is deleted.

```json
{
  "type": "message.delete",
  "payload": {
    "id": 123,
    "chat_id": 1
  }
}
```

---

#### message.read

Received when another user reads messages in your chat.

```json
{
  "type": "message.read",
  "payload": {
    "chat_id": 1,
    "user_id": 2,
    "message_id": 150,
    "read_at": "2025-01-15T14:40:00Z"
  }
}
```

---

#### typing.start

Received when another user starts typing in a chat.

```json
{
  "type": "typing.start",
  "payload": {
    "chat_id": 1,
    "user_id": 2
  }
}
```

---

#### typing.stop

Received when another user stops typing.

```json
{
  "type": "typing.stop",
  "payload": {
    "chat_id": 1,
    "user_id": 2
  }
}
```

---

#### presence.online

Received when a contact comes online.

```json
{
  "type": "presence.online",
  "payload": {
    "user_id": 2,
    "online": true
  }
}
```

---

#### presence.offline

Received when a contact goes offline.

```json
{
  "type": "presence.offline",
  "payload": {
    "user_id": 2,
    "online": false
  }
}
```

---

#### error

Received when an error occurs processing a client message.

```json
{
  "type": "error",
  "payload": {
    "code": "invalid_chat",
    "message": "You are not a participant of this chat"
  }
}
```

---

### Client Events (Client to Server)

#### typing.start

Send when the user starts typing in a chat.

```json
{
  "type": "typing.start",
  "payload": {
    "chat_id": 1
  }
}
```

---

#### typing.stop

Send when the user stops typing (e.g., after a timeout or clearing the input).

```json
{
  "type": "typing.stop",
  "payload": {
    "chat_id": 1
  }
}
```

---

### Connection Lifecycle

1. **Connect:** Client establishes WebSocket connection with token
2. **Authenticate:** Server validates token and retrieves user info
3. **Subscribe:** Client is auto-subscribed to all their chats
4. **Presence:** Server broadcasts `presence.online` to user's contacts
5. **Active:** Client receives events and can send typing indicators
6. **Disconnect:** Server broadcasts `presence.offline` and cleans up

---

### Reconnection Strategy

Clients should implement automatic reconnection with exponential backoff:

```javascript
const connect = (attempt = 0) => {
  const ws = new WebSocket(`ws://localhost:9900/chat/ws?token=${token}`);

  ws.onclose = () => {
    const delay = Math.min(1000 * Math.pow(2, attempt), 30000);
    setTimeout(() => connect(attempt + 1), delay);
  };

  ws.onopen = () => {
    attempt = 0; // Reset on successful connection
  };
};
```

---

### JavaScript Client Example

```javascript
class ChatWSClient {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.handlers = {};
  }

  connect() {
    this.ws = new WebSocket(`ws://localhost:9900/chat/ws?token=${this.token}`);

    this.ws.onmessage = (event) => {
      const { type, payload } = JSON.parse(event.data);
      if (this.handlers[type]) {
        this.handlers[type](payload);
      }
    };

    this.ws.onclose = () => {
      // Implement reconnection logic
      setTimeout(() => this.connect(), 3000);
    };
  }

  on(eventType, handler) {
    this.handlers[eventType] = handler;
  }

  sendTyping(chatId, isTyping) {
    this.ws.send(JSON.stringify({
      type: isTyping ? 'typing.start' : 'typing.stop',
      payload: { chat_id: chatId }
    }));
  }

  close() {
    this.ws.close();
  }
}

// Usage
const client = new ChatWSClient(accessToken);
client.connect();

client.on('message.new', (payload) => {
  console.log('New message:', payload);
});

client.on('typing.start', (payload) => {
  console.log(`User ${payload.user_id} is typing in chat ${payload.chat_id}`);
});

client.on('presence.online', (payload) => {
  console.log(`User ${payload.user_id} came online`);
});
```

---

### WebSocket Data Models

#### WebSocket Message Payload

```typescript
interface MessagePayload {
  id: number;
  chat_id: number;
  sender_id: number;
  content: string;
  sent_at?: string;     // RFC3339 timestamp
  edited_at?: string;   // RFC3339 timestamp, only for edits
}
```

#### Message Delete Payload

```typescript
interface MessageDeletePayload {
  id: number;
  chat_id: number;
}
```

#### Message Read Payload

```typescript
interface MessageReadPayload {
  chat_id: number;
  user_id: number;
  message_id: number;
  read_at: string;      // RFC3339 timestamp
}
```

#### Typing Payload

```typescript
interface TypingPayload {
  chat_id: number;
  user_id: number;
}
```

#### Presence Payload

```typescript
interface PresencePayload {
  user_id: number;
  online: boolean;
  last_seen?: string;   // RFC3339 timestamp, when offline
}
```

#### Error Payload

```typescript
interface ErrorPayload {
  code: string;
  message: string;
}
```

---

## Data Models

### User

```typescript
interface User {
  user_id: number;
  username: string; // 3-30 chars, alphanumeric + underscore
  email: string; // Valid email format
  role: "user" | "admin";
  image_path: string | null;
  created_at: string; // RFC3339 timestamp
}
```

### Direct Message List Item

```typescript
interface DMListItem {
  chat_id: number;
  other_user_id: number;
  other_username: string;
  other_user_image: string | null;
  last_message_text: string | null;
  last_message_sent_at: string | null;
  unread_count: number;
}
```

### Group List Item

```typescript
interface GroupListItem {
  chat_id: number;
  name: string;
  creator_id: number;
  participant_count: number;
  last_message_text: string | null;
  last_message_sent_at: string | null;
  unread_count: number;
}
```

### Chat Detail

```typescript
interface Chat {
  chat_id: number;
  type: "direct" | "group";
  name: string; // Empty for DMs
  creator_id: number; // 0 for DMs
  participants: ChatParticipant[];
  created_at: string;
}

interface ChatParticipant {
  user_id: number;
  username: string;
  image_path: string | null;
  joined_at: string;
}
```

### Message

```typescript
interface Message {
  message_id: number;
  chat_id: number;
  sender_id: number;
  sender_name: string;
  sender_image: string | null;
  content: string;
  sent_at: string;
  edited_at: string | null;
}
```

### User Online Status

```typescript
interface UserOnlineStatus {
  user_id: number;
  is_online: boolean;
  last_seen: string | null; // null if currently online
}
```

---

## Notes for Frontend Implementation

### Authentication Flow

1. Call `POST /auth/login` to get tokens
2. Store `access_token` and `refresh_token` securely (localStorage/sessionStorage)
3. Include `Authorization: Bearer <access_token>` header in all authenticated requests
4. When access token expires (401 response), implement token refresh logic
5. Call `POST /auth/logout` on user logout

### Real-time Features

ChatX provides WebSocket support for real-time messaging. See the [WebSocket API](#websocket-api) section for details.

**Recommended implementation:**

1. Connect to WebSocket on app load: `ws://localhost:9900/chat/ws?token=<access_token>`
2. Listen for events: `message.new`, `message.edit`, `message.delete`, `typing.*`, `presence.*`
3. Send typing indicators when user types in chat input
4. Implement reconnection with exponential backoff
5. Fall back to polling if WebSocket is unavailable

### Image Upload Flow

1. Call `POST /auth/images/upload` with the image file (multipart/form-data)
2. Receive `image_path` in the response
3. Call `PUT /auth/users/me/image` with the returned `image_path` to update your profile

### Pagination Best Practices

- Start with `page=0`
- Use consistent `limit` values (20-50 recommended)
- Check `total` to calculate total pages: `Math.ceil(total / limit)`
- Display "Load More" or pagination controls based on `total`

### Error Handling

- Always check response status codes
- Parse error responses for field-specific validation errors
- Handle 401 by redirecting to login
- Handle 403 by showing "insufficient permissions" message
- Handle 404 by showing "not found" message
- Handle 500 by showing generic error message

---

## API Testing

### Example cURL Commands

**Login:**

```bash
curl -X POST http://localhost:9900/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Get Current User:**

```bash
curl -X GET http://localhost:9900/auth/users/me \
  -H "Authorization: Bearer <your_token>"
```

**Get DM List:**

```bash
curl -X GET "http://localhost:9900/chat/chats/dms?page=0&limit=20" \
  -H "Authorization: Bearer <your_token>"
```

**Send Message:**

```bash
curl -X POST http://localhost:9900/chat/messages \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{"chat_id":1,"content":"Hello!"}'
```

**Get Chat Messages:**

```bash
curl -X GET "http://localhost:9900/chat/chats/1/messages?page=0&limit=50" \
  -H "Authorization: Bearer <your_token>"
```

---

## Quick Reference: All Endpoints

### Authentication & Users

| Method | Endpoint                | Auth  | Description          |
| ------ | ----------------------- | ----- | -------------------- |
| POST   | /auth/login             | No    | Login                |
| POST   | /auth/logout            | Yes   | Logout               |
| POST   | /auth/users             | Admin | Create user          |
| GET    | /auth/users             | Admin | List users           |
| GET    | /auth/users/{user_id}   | Admin | Get user details     |
| DELETE | /auth/users/{user_id}   | Admin | Delete user          |
| GET    | /auth/users/me          | Yes   | Get current user     |
| PUT    | /auth/users/me/password | Yes   | Change password      |
| PUT    | /auth/users/me/image    | Yes   | Update profile image |

### Images

| Method | Endpoint                   | Auth | Description   |
| ------ | -------------------------- | ---- | ------------- |
| POST   | /auth/images/upload        | Yes  | Upload image  |
| GET    | /auth/images/{image_path...} | No   | Download image |

### Chats

| Method | Endpoint              | Auth | Description           |
| ------ | --------------------- | ---- | --------------------- |
| GET    | /chat/chats/dms       | Yes  | List DM conversations |
| GET    | /chat/chats/groups    | Yes  | List group chats      |
| GET    | /chat/chats/{chat_id} | Yes  | Get chat details      |
| POST   | /chat/chats/dms       | Yes  | Create DM             |
| POST   | /chat/chats/groups    | Yes  | Create group chat     |

### Messages

| Method | Endpoint                       | Auth | Description    |
| ------ | ------------------------------ | ---- | -------------- |
| GET    | /chat/chats/{chat_id}/messages | Yes  | List messages  |
| POST   | /chat/messages                 | Yes  | Send message   |
| PUT    | /chat/messages/{message_id}    | Yes  | Edit message   |
| DELETE | /chat/messages/{message_id}    | Yes  | Delete message |

### Notifications

| Method | Endpoint                     | Auth | Description             |
| ------ | ---------------------------- | ---- | ----------------------- |
| GET    | /chat/notifications/unread   | Yes  | Total unread count      |
| GET    | /chat/chats/{chat_id}/unread | Yes  | Chat unread count       |
| POST   | /chat/chats/read             | Yes  | Mark messages as read   |
| POST   | /chat/users/online-status    | Yes  | Get users online status |

### WebSocket

| Method | Endpoint   | Auth | Description                  |
| ------ | ---------- | ---- | ---------------------------- |
| GET    | /chat/ws   | Yes* | WebSocket connection (*token via query param) |

---
