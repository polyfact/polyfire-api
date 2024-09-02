# Polyfire API Documentation

Welcome to the Polyfire API documentation. This guide will help you understand how to interact with the Polyfire API, which is hosted at `https://api.polyfire.com`.

## Base URL

All API endpoints referenced in this documentation have the following base URL:

```
https://api.polyfire.com
```

## Authentication

Polyfire API uses token-based authentication. You need to include your token in the `Authorization` header of your requests.

## Endpoints

### Auth Routes

#### Exchange Firebase Token
- **Endpoint:** `GET /project/:id/auth/firebase`
- **Description:** Exchange a Firebase token for a Polyfire token.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/project/{id}/auth/firebase" -H "Authorization: Bearer {token}"
  ```

#### Exchange Custom Token
- **Endpoint:** `GET /project/:id/auth/custom`
- **Description:** Exchange a custom token for a Polyfire token.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/project/{id}/auth/custom" -H "Authorization: Bearer {token}"
  ```

#### Anonymous Token Exchange
- **Endpoint:** `GET /project/:id/auth/anonymous`
- **Description:** Exchange an anonymous token for a Polyfire token.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/project/{id}/auth/anonymous" -H "Authorization: Bearer {token}"
  ```

#### Provider Redirect
- **Endpoint:** `GET /project/:id/auth/provider/redirect`
- **Description:** Redirect to an external provider for authentication.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/project/{id}/auth/provider/redirect" -H "Authorization: Bearer {token}"
  ```

#### Provider Callback
- **Endpoint:** `GET /project/:id/auth/provider/callback`
- **Description:** Handle the callback from an external provider.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/project/{id}/auth/provider/callback" -H "Authorization: Bearer {token}"
  ```

#### Refresh Token
- **Endpoint:** `POST /project/:id/auth/provider/refresh`
- **Description:** Refresh an authentication token.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/project/{id}/auth/provider/refresh" -H "Authorization: Bearer {token}"
  ```

#### Get Auth ID
- **Endpoint:** `GET /auth/id`
- **Description:** Retrieve the authenticated user's ID.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/auth/id" -H "Authorization: Bearer {token}"
  ```

### Usage Routes

#### Get Usage
- **Endpoint:** `GET /usage`
- **Description:** Retrieve the usage statistics for the authenticated user.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/usage" -H "Authorization: Bearer {token}"
  ```

### Completion Routes

#### Generate
- **Endpoint:** `POST /generate`
- **Description:** Generate a completion.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/generate" -H "Authorization: Bearer {token}" -d '{ "input": "Your input here" }'
  ```

#### Get Chat History
- **Endpoint:** `GET /chat/:id/history`
- **Description:** Retrieve the history of a chat.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/chat/{id}/history" -H "Authorization: Bearer {token}"
  ```

#### List Chats
- **Endpoint:** `GET /chats`
- **Description:** List all chats.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/chats" -H "Authorization: Bearer {token}"
  ```

#### Create Chat
- **Endpoint:** `POST /chats`
- **Description:** Create a new chat.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/chats" -H "Authorization: Bearer {token}" -d '{ "name": "New Chat" }'
  ```

#### Update Chat
- **Endpoint:** `PUT /chat/:id`
- **Description:** Update an existing chat.
- **Example Request:**
  ```bash
  curl -X PUT "https://api.polyfire.com/chat/{id}" -H "Authorization: Bearer {token}" -d '{ "name": "Updated Chat" }'
  ```

#### Delete Chat
- **Endpoint:** `DELETE /chat/:id`
- **Description:** Delete a chat.
- **Example Request:**
  ```bash
  curl -X DELETE "https://api.polyfire.com/chat/{id}" -H "Authorization: Bearer {token}"
  ```

#### Stream
- **Endpoint:** `GET /stream`
- **Description:** Stream a completion.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/stream" -H "Authorization: Bearer {token}"
  ```

### Transcription Routes

#### Transcribe
- **Endpoint:** `POST /transcribe`
- **Description:** Transcribe audio to text.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/transcribe" -H "Authorization: Bearer {token}" -d '{ "audio": "base64_encoded_audio" }'
  ```

### TTS Routes

#### Text to Speech
- **Endpoint:** `POST /tts`
- **Description:** Convert text to speech.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/tts" -H "Authorization: Bearer {token}" -d '{ "text": "Hello, world!" }'
  ```

### Image Generation Routes

#### Generate Image
- **Endpoint:** `GET /image/generate`
- **Description:** Generate an image.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/image/generate" -H "Authorization: Bearer {token}" -d '{ "prompt": "A sunset over the mountains" }'
  ```

### Memory Routes

#### List Memories
- **Endpoint:** `GET /memories`
- **Description:** List all memories.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/memories" -H "Authorization: Bearer {token}"
  ```

#### Search Memory
- **Endpoint:** `POST /memory/:id/search`
- **Description:** Search within a memory.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/memory/{id}/search" -H "Authorization: Bearer {token}" -d '{ "query": "search term" }'
  ```

#### Create Memory
- **Endpoint:** `POST /memory`
- **Description:** Create a new memory.
- **Example Request:**
  ```bash
  curl -X POST "https://api.polyfire.com/memory" -H "Authorization: Bearer {token}" -d '{ "data": "memory data" }'
  ```

#### Add to Memory
- **Endpoint:** `PUT /memory`
- **Description:** Add data to an existing memory.
- **Example Request:**
  ```bash
  curl -X PUT "https://api.polyfire.com/memory" -H "Authorization: Bearer {token}" -d '{ "data": "additional memory data" }'
  ```

### KV Routes

#### Get KV
- **Endpoint:** `GET /kv`
- **Description:** Retrieve a key-value pair.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/kv" -H "Authorization: Bearer {token}" -d '{ "key": "example_key" }'
  ```

#### List KVs
- **Endpoint:** `GET /kvs`
- **Description:** List all key-value pairs.
- **Example Request:**
  ```bash
  curl -X GET "https://api.polyfire.com/kvs" -H "Authorization: Bearer {token}"
  ```

#### Set KV
- **Endpoint:** `PUT /kv`
- **Description:** Set a key-value pair.
- **Example Request:**
  ```bash
  curl -X PUT "https://api.polyfire.com/kv" -H "Authorization: Bearer {token}" -d '{ "key": "example_key", "value": "example_value" }'
  ```

#### Delete KV
- **Endpoint:** `DELETE /kv`
- **Description:** Delete a key-value pair.
- **Example Request:**
  ```bash
  curl -X DELETE "https://api.polyfire.com/kv" -H "Authorization: Bearer {token}" -d '{ "key": "example_key" }'
  ```

## Error Handling

The Polyfire API uses standard HTTP status codes to indicate the success or failure of an API request. Here are some common status codes you might encounter:

- `200 OK`: The request was successful.
- `400 Bad Request`: The request was invalid or cannot be otherwise served.
- `401 Unauthorized`: Authentication failed or user does not have permissions for the requested operation.
- `404 Not Found`: The requested resource could not be found.
- `500 Internal Server Error`: An error occurred on the server.

## Conclusion

This documentation provides an overview of the available endpoints and their usage. For more detailed information, please refer to the specific endpoint documentation or contact our support team.