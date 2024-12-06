# AnyQA API Documentation

## Overview

AnyQA is a presentation assistant service that provides bilingual support for Chinese speakers delivering English presentations. The API supports real-time question handling, session management, and WebSocket connections for live interactions.

## Base URL

The API base URL depends on your server configuration, defined in the config file.

## Authentication

CORS is enabled for all origins. No authentication is currently required for API endpoints.

## API Endpoints

### Submit a Question

`POST /api/question`

Submit a new question to get AI-powered presentation assistance.

#### Request Body

```json
{
    "sessionId": "string",
    "content": "string"
}
```

- `sessionId`: Unique identifier for the presentation session
- `content`: The question or content to be processed

#### Response

```json
{
    "status": "success"
}
```

#### Error Response

```json
{
    "error": "error message"
}
```

### Get Session Questions

`GET /api/questions/:sessionId`

Retrieve all questions for a specific session.

#### Parameters

- `sessionId`: Session identifier (URL parameter)

#### Response

```json
[
    {
        "id": "integer",
        "content": "string",
        "status": "string",
        "ai_suggestion": "string",
        "created_at": "string"
    }
]
```

### Update Question Status

`POST /api/question/status`

Update the status of a specific question.

#### Request Body

```json
{
    "id": "integer",
    "status": "string"
}
```

- `id`: Question identifier
- `status`: New status value (e.g., "showing", "finished")

Note: When setting status to "showing", all other questions with "showing" status will be automatically set to "finished".

#### Response

```json
{
    "status": "success"
}
```

### Delete Question

`DELETE /api/question/:id`

Delete a specific question.

#### Parameters

- `id`: Question identifier (URL parameter)

#### Response

Empty response with status code 200 on success.

### WebSocket Connection

`GET /api/ws`

Establish a WebSocket connection for real-time communication.

#### WebSocket Events

- The server echoes back any message received through the WebSocket connection.

## Response Formats

### Question Object

```json
{
    "id": "integer",
    "content": "string",
    "status": "string",
    "ai_suggestion": "string",
    "created_at": "string"
}
```

- `id`: Unique identifier for the question
- `content`: Original question content
- `status`: Current status of the question (e.g., "showing", "finished")
- `ai_suggestion`: AI-generated response (may be null)
- `created_at`: Timestamp of question creation

## Error Handling

All endpoints may return the following error responses:

- 400 Bad Request: Invalid request format or parameters
- 500 Internal Server Error: Server-side processing error

Error responses include an "error" field with a description of the error:

```json
{
    "error": "error description"
}
```

## AI Response Format

The AI responses are structured in markdown format with two main sections:

1. Chinese Understanding
   - Core message
   - Key terms (Chinese and English)
   - Key points

2. English Delivery
   - Quick answer
   - Key speaking points
   - Example (marked as real or hypothetical)
   - Useful phrases

## Notes

- All API endpoints support CORS
- The server uses MySQL for data persistence
- AI responses are generated using OpenAI's API
- WebSocket support is included for real-time communication
- Status updates maintain only one "showing" status at a time