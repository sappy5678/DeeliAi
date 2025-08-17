# DeeliAi API Documentation

This document provides an overview and detailed specifications for the DeeliAi API. All API endpoints are prefixed with `/api/v1`.

## Authentication

Most endpoints require authentication using a Bearer Token. After successful login, you will receive a JWT token that should be included in the `Authorization` header of your requests as `Bearer <YOUR_TOKEN>`.

## Error Handling

API errors are returned with a JSON body containing the following structure:

```json
{
  "name": "string",
  "code": "integer",
  "message": "string",
  "remoteCode": "integer",
  "detail": {}
}
```

Common error codes include:
- `400 Bad Request`: Invalid parameters or request body.
- `401 Unauthorized`: Authentication failed or token is missing/invalid.
- `404 Not Found`: Resource not found.

## Endpoints

### Health Check

#### `GET /health`

*   **Summary:** Health check endpoint.
*   **Responses:**
    *   `200 OK`

### User Management

#### `POST /user/signup`

*   **Summary:** Register a new user.
*   **Request Body:**
    ```json
    {
      "email": "string" (email format),
      "username": "string",
      "password": "string"
    }
    ```
*   **Responses:**
    *   `201 Created`:
        ```json
        {
          "user": {
            "id": "string" (uuid),
            "email": "string" (email format),
            "username": "string"
          },
          "token": "string"
        }
        ```
    *   `400 Bad Request`: Invalid parameters.

#### `POST /user/login`

*   **Summary:** Authenticate a user and receive an access token.
*   **Request Body:**
    ```json
    {
      "email": "string" (email format),
      "password": "string"
    }
    ```
*   **Responses:**
    *   `200 OK`:
        ```json
        {
          "user": {
            "id": "string" (uuid),
            "email": "string" (email format),
            "username": "string"
          },
          "token": "string"
        }
        ```
    *   `400 Bad Request`: Invalid parameters.
    *   `401 Unauthorized`: Invalid credentials.

#### `GET /user/me`

*   **Summary:** Get the current authenticated user's information.
*   **Security:** Bearer Token required.
*   **Responses:**
    *   `200 OK`:
        ```json
        {
          "id": "string" (uuid),
          "email": "string" (email format),
          "username": "string"
        }
        ```
    *   `401 Unauthorized`: Authentication failed.

### Article Management

#### `POST /articles`

*   **Summary:** Create a new article.
*   **Security:** Bearer Token required.
*   **Request Body:**
    ```json
    {
      "url": "string" (url format)
    }
    ```
*   **Responses:**
    *   `201 Created`:
        ```json
        {
          "id": "string" (uuid),
          "url": "string"
        }
        ```
    *   `400 Bad Request`: Invalid parameters.
    *   `401 Unauthorized`: Authentication failed.

#### `GET /articles`

*   **Summary:** List articles for the current user.
*   **Security:** Bearer Token required.
*   **Query Parameters:**
    *   `after` (string, uuid): Article ID to start listing after (for pagination).
    *   `limit` (integer, default: 10): Maximum number of articles to return.
*   **Responses:**
    *   `200 OK`:
        ```json
        {
          "articles": [
            {
              "id": "string" (uuid),
              "url": "string",
              "title": "string" (optional),
              "description": "string" (optional),
              "image_url": "string" (optional)
            }
          ]
        }
        ```
    *   `400 Bad Request`: Invalid query parameters.
    *   `401 Unauthorized`: Authentication failed.

#### `DELETE /articles/{article_id}`

*   **Summary:** Delete an article.
*   **Security:** Bearer Token required.
*   **Path Parameters:**
    *   `article_id` (string, uuid): ID of the article to delete.
*   **Responses:**
    *   `204 No Content`
    *   `401 Unauthorized`: Authentication failed.
    *   `404 Not Found`: Article not found.

#### `PUT /articles/{article_id}/rate`

*   **Summary:** Rate an article.
*   **Security:** Bearer Token required.
*   **Path Parameters:**
    *   `article_id` (string, uuid): ID of the article to rate.
*   **Request Body:**
    ```json
    {
      "rate": "integer" (int)
    }
    ```
*   **Responses:**
    *   `204 No Content`
    *   `400 Bad Request`: Invalid parameters.
    *   `401 Unauthorized`: Authentication failed.

#### `GET /articles/{article_id}/rate`

*   **Summary:** Get the rating of an article.
*   **Security:** Bearer Token required.
*   **Path Parameters:**
    *   `article_id` (string, uuid): ID of the article to get rating for.
*   **Responses:**
    *   `200 OK`:
        ```json
        {
          "rate": "integer" (int)
        }
        ```
    *   `401 Unauthorized`: Authentication failed.
    *   `404 Not Found`: Article or rating not found.

#### `DELETE /articles/{article_id}/rate`

*   **Summary:** Delete the rating of an article.
*   **Security:** Bearer Token required.
*   **Path Parameters:**
    *   `article_id` (string, uuid): ID of the article to delete rating for.
*   **Responses:**
    *   `204 No Content`
    *   `401 Unauthorized`: Authentication failed.

#### `GET /articles/recommendations`

*   **Summary:** Get article recommendations.
*   **Security:** Bearer Token required.
*   **Query Parameters:**
    *   `limit` (integer, default: 10, min: 1, max: 50): Maximum number of recommendations to return.
*   **Responses:**
    *   `200 OK`:
        ```json
        {
          "items": [
            {
              "article": {
                "id": "string" (uuid),
                "url": "string",
                "title": "string" (optional),
                "description": "string" (optional),
                "image_url": "string" (optional)
              },
              "average_rating": "number" (double)
            }
          ]
        }
        ```
    *   `400 Bad Request`: Invalid query parameters.
    *   `401 Unauthorized`: Authentication failed.
