# MeGo (RestApi-Go)

A lightweight and efficient REST API backend built with Go, featuring user management, JWT authentication, and PostgreSQL integration.

## Features

- RESTful API architecture using the Chi router.
- Secure user registration and login with password hashing.
- JWT-based authentication using HTTP-only cookies.
- Type-safe database interactions with SQLC.
- PostgreSQL database integration.
- Environment variable management with godotenv.
- Graceful server shutdown and health checks.
- Comprehensive unit testing suite.

## Tech Stack

- Language: Go (1.25.3+)
- Router: Chi
- Database: PostgreSQL
- Query Generator: SQLC
- Authentication: JWT (github.com/golang-jwt/jwt/v4)
- Others: godotenv, uuid, crypto/bcrypt

## Prerequisites

- Go 1.25.3 or higher.
- PostgreSQL database instance.
- SQLC (optional, for regenerating database code).

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Blue-Onion/RestApi-Go.git
   cd MeGo
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## Configuration

Create a `.env` file in the root directory and configure the following variables:

```env
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
JWT_SECERT=your_super_secret_key
```

## Running the Application

You can start the server using the provided Makefile or directly with Go:

Using Makefile:
```bash
make run
```

Using Go:
```bash
go run cmd/main.go
```

The server will start listening on the port specified in your `.env` file (default: http://localhost:8080).

## API Reference & curl Testing Guide (LLM-Friendly)

All JSON API responses conform to a unified wrapper structure:

### Success Response Format
```json
{
  "Success": true,
  "Data": <payload_object_or_array_or_string>
}
```

### Error Response Format
```json
{
  "Success": false,
  "Data": {
    "Error": "error message description"
  }
}
```

### Authentication System
Authentication uses a JWT token stored in an HTTP-only cookie named `authToken`. 
To test authenticated endpoints using `curl`, you must:
1. Save the cookie upon a successful login request using the `-c cookies.txt` flag.
2. Send the saved cookie in subsequent requests using the `-b cookies.txt` flag.

---

### Endpoint Directory

#### 1. Public & Core Endpoints

* **Get Home Page** (Returns HTML)
  * `GET /`
  * Auth: No
  * curl:
    ```bash
    curl -X GET http://localhost:8080/
    ```

* **Health Check**
  * `GET /health`
  * Auth: No
  * curl:
    ```bash
    curl -X GET http://localhost:8080/health
    ```

#### 2. User & Authentication (`/auth`)

* **Register a New User**
  * `POST /auth/users`
  * Auth: No
  * Content-Type: `application/json`
  * Body Parameters:
    * `name` (string, required)
    * `email` (string, required)
    * `password` (string, required)
    * `description` (string, optional)
    * `batch` (string, required)
    * `social` (JSON object, optional, e.g., `{"twitter": "@handle"}`)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/auth/users \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Jane Doe",
        "email": "jane@example.com",
        "password": "mypassword123",
        "description": "Artist and sculptor",
        "batch": "2026",
        "social": {
          "instagram": "@jane_art",
          "portfolio": "https://jane.art"
        }
      }'
    ```

* **User Login (Saves auth cookie)**
  * `POST /auth/login`
  * Auth: No
  * Content-Type: `application/json`
  * Body Parameters:
    * `email` (string, required)
    * `password` (string, required)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/auth/login \
      -H "Content-Type: application/json" \
      -c cookies.txt \
      -d '{
        "email": "jane@example.com",
        "password": "mypassword123"
      }'
    ```

* **User Logout (Clears auth cookie)**
  * `POST /auth/logout`
  * Auth: Yes
  * curl:
    ```bash
    curl -X POST http://localhost:8080/auth/logout \
      -b cookies.txt
    ```

* **Upload/Update Profile & Banner Images**
  * `PATCH /auth/users/avatar`
  * Auth: Yes
  * Content-Type: `multipart/form-data`
  * Form Fields:
    * `user_image` (file, optional)
    * `banner_image` (file, optional)
    *(At least one of these two files must be included)*
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/auth/users/avatar \
      -b cookies.txt \
      -F "user_image=@/path/to/avatar.jpg" \
      -F "banner_image=@/path/to/banner.jpg"
    ```

* **Update User Profile Details**
  * `PATCH /auth/users/{id}`
  * Auth: Yes (Owner or Admin role required)
  * Path Param: `{id}` (UUID, required)
  * Content-Type: `application/json`
  * Body Parameters (all fields optional):
    * `name` (string)
    * `email` (string)
    * `batch` (string)
    * `description` (string)
    * `social` (JSON object)
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/auth/users/<USER_UUID> \
      -b cookies.txt \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Jane Smith",
        "description": "Master painter"
      }'
    ```

* **Change Password**
  * `PATCH /auth/users/password`
  * Auth: Yes
  * Content-Type: `application/json`
  * Body Parameters:
    * `old_password` (string, required)
    * `password` (string, required - the new password)
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/auth/users/password \
      -b cookies.txt \
      -H "Content-Type: application/json" \
      -d '{
        "old_password": "mypassword123",
        "password": "newsecurepassword456"
      }'
    ```

#### 3. Arts & Social Interactions (`/art`)

* **Get Arts by User**
  * `GET /art/user/{user_id}`
  * Auth: No
  * Path Param: `{user_id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/art/user/<USER_UUID>
    ```

* **Get Art by ID**
  * `GET /art/{id}`
  * Auth: No
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/art/<ART_UUID>
    ```

* **Get Comments on Art**
  * `GET /art/{id}/comments`
  * Auth: No
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/art/<ART_UUID>/comments
    ```

* **Get Comment Count**
  * `GET /art/{id}/comments/count`
  * Auth: No
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/art/<ART_UUID>/comments/count
    ```

* **Get Likes Count**
  * `GET /art/{id}/likes/count`
  * Auth: No
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/art/<ART_UUID>/likes/count
    ```

* **Create New Art**
  * `POST /art/`
  * Auth: Yes
  * Content-Type: `multipart/form-data`
  * Form Fields:
    * `image` (file, required, max 5MB)
    * `name` (string, required, length >= 3)
    * `description` (string, optional)
    * `tags` (repeated string array, optional, can pass multiple `tags` fields)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/art/ \
      -b cookies.txt \
      -F "image=@/path/to/artwork.png" \
      -F "name=Starry Night Refraction" \
      -F "description=My latest canvas work" \
      -F "tags=impressionism" \
      -F "tags=oil"
    ```

* **Update Art Metadata**
  * `PATCH /art/{id}`
  * Auth: Yes (Owner of the art required)
  * Path Param: `{id}` (UUID, required)
  * Content-Type: `multipart/form-data` or form values
  * Form Fields:
    * `name` (string, required, length >= 3)
    * `description` (string, optional)
    * `tags` (repeated string array, optional)
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/art/<ART_UUID> \
      -b cookies.txt \
      -F "name=Starry Night 2.0" \
      -F "description=An updated view" \
      -F "tags=modern"
    ```

* **Delete Art**
  * `DELETE /art/{id}`
  * Auth: Yes (Owner of the art required)
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X DELETE http://localhost:8080/art/<ART_UUID> \
      -b cookies.txt
    ```

* **Add Comment to Art**
  * `POST /art/{art_id}/comment`
  * Auth: Yes
  * Path Param: `{art_id}` (UUID, required)
  * Content-Type: `application/json`
  * Body Parameters:
    * `comment` (string, required)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/art/<ART_UUID>/comment \
      -b cookies.txt \
      -H "Content-Type: application/json" \
      -d '{
        "comment": "Absolutely brilliant use of light!"
      }'
    ```

* **Delete Comment**
  * `DELETE /art/comment/{id}`
  * Auth: Yes (Owner of the comment required)
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X DELETE http://localhost:8080/art/comment/<COMMENT_UUID> \
      -b cookies.txt
    ```

* **Like an Art**
  * `POST /art/{art_id}/like`
  * Auth: Yes
  * Path Param: `{art_id}` (UUID, required)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/art/<ART_UUID>/like \
      -b cookies.txt
    ```

* **Unlike an Art**
  * `POST /art/{art_id}/unlike`
  * Auth: Yes
  * Path Param: `{art_id}` (UUID, required)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/art/<ART_UUID>/unlike \
      -b cookies.txt
    ```

#### 4. Events & Attendance (`/event`)

* **List All Events**
  * `GET /event/`
  * Auth: No
  * curl:
    ```bash
    curl -X GET http://localhost:8080/event/
    ```

* **Get Event by ID**
  * `GET /event/{id}`
  * Auth: No
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/event/<EVENT_UUID>
    ```

* **Join Event**
  * `POST /event/{id}/join`
  * Auth: Yes
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X POST http://localhost:8080/event/<EVENT_UUID>/join \
      -b cookies.txt
    ```

* **Leave/Remove Event Attendee**
  * `DELETE /event/{id}/attendee`
  * Auth: Yes
  * Path Param: `{id}` (UUID, required)
  * Query Param: `user_id` (UUID, required)
  * curl:
    ```bash
    curl -X DELETE "http://localhost:8080/event/<EVENT_UUID>/attendee?user_id=<USER_UUID>" \
      -b cookies.txt
    ```

* **List Event Attendees**
  * `GET /event/{id}/attendees`
  * Auth: Yes
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X GET http://localhost:8080/event/<EVENT_UUID>/attendees \
      -b cookies.txt
    ```

* **Create Event (Admin Only)**
  * `POST /event/`
  * Auth: Yes (Admin role required)
  * Content-Type: `multipart/form-data`
  * Form Fields:
    * `name` (string, required, length >= 3)
    * `date` (string, required, format: `YYYY-MM-DD`)
    * `description` (string, optional)
    * `venue` (string, optional)
    * `status` (string, required, enum: `online`, `offline`)
    * `image` (file, optional)
    * `banner_image` (file, optional)
    *(At least one image or banner_image must be provided)*
  * curl:
    ```bash
    curl -X POST http://localhost:8080/event/ \
      -b cookies.txt \
      -F "name=Annual Art Gala" \
      -F "date=2026-06-15" \
      -F "description=Exhibition and live auction" \
      -F "venue=Grand Hall Room B" \
      -F "status=offline" \
      -F "image=@/path/to/event_logo.png"
    ```

* **Update Event (Admin Only)**
  * `PATCH /event/{id}`
  * Auth: Yes (Admin role required)
  * Path Param: `{id}` (UUID, required)
  * Content-Type: `multipart/form-data`
  * Form Fields: Same as Create Event (Name, Date, Description, Venue, Status, Image, BannerImage)
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/event/<EVENT_UUID> \
      -b cookies.txt \
      -F "name=Annual Art Gala 2026" \
      -F "date=2026-06-16" \
      -F "status=online" \
      -F "image=@/path/to/new_logo.png"
    ```

* **Delete Event (Admin Only)**
  * `DELETE /event/{id}`
  * Auth: Yes (Admin role required)
  * Path Param: `{id}` (UUID, required)
  * curl:
    ```bash
    curl -X DELETE http://localhost:8080/event/<EVENT_UUID> \
      -b cookies.txt
    ```

#### 5. Admin Actions (`/admin`)

* **Moderate User Status / Role (Admin Only)**
  * `PATCH /admin/users/{id}/status`
  * Auth: Yes (Admin role required)
  * Path Param: `{id}` (UUID, required)
  * Content-Type: `application/json`
  * Body Parameters (Role or Status must be provided; both cannot be empty):
    * `role` (string, optional, enum: `user`, `admin`)
    * `status` (string, optional, enum: `pending`, `approved`, `banned`)
  * curl:
    ```bash
    curl -X PATCH http://localhost:8080/admin/users/<USER_UUID>/status \
      -b cookies.txt \
      -H "Content-Type: application/json" \
      -d '{
        "status": "approved"
      }'
    ```

* **Moderate Art Status (Admin Only)**
  * `PATCH /admin/arts/{art_id}/status`
  * Auth: Yes (Admin role required)
  * Path Param: `{art_id}` (UUID, required)
  * Query Param: `status` (string, required, enum: `approved`, `pending`, `rejected` / `banned`)
  * curl:
    ```bash
    curl -X PATCH "http://localhost:8080/admin/arts/<ART_UUID>/status?status=approved" \
      -b cookies.txt
    ```


## Project Structure

- `cmd/`: Application entry point.
- `config/`: Configuration loading and database connection.
- `handler/`: HTTP handlers for various routes.
- `internal/database/`: Auto-generated database code by SQLC.
- `middleware/`: Authentication and other middlewares.
- `model/`: Data models and structures.
- `sql/`: SQL schema and queries.
- `test/`: Automated test suite.
- `utlis/`: Utility functions (JWT, hashing, etc.).

## Testing

To run the automated tests, use the following command:

```bash
go test ./test/...
```
