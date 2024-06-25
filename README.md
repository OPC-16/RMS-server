# Go-based RMS-server Application

Backend server for a Recruitment Management System

## Features

1. **Code Quality**: The codebase adheres to clean code practices, emphasizing modularity and readability.
2. **Redis Database**: Uses Redis as a database for storing user information and logs.
3. **JWT Token Authentication**: Uses JWT for user authentication.
4. **Middleware**: Utilizes middleware to enhance functionality, such as JWT authentication and role-based access control.

## Installation

### Prerequisites

- Go 1.16 or higher
- Redis

### Steps

1. Clone the repository:
    ```sh
    git clone https://github.com/OPC-16/RMS-server
    cd RMS-server
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Run Redis server (if not already running):
    ```sh
    redis-server
    ```
    or
   ```bash
   docker run -p 6379:6379 redis:latest
   ```

5. Build and run the application:
    ```sh
    go run main.go
    ```

## Usage

### Echo Server

This project uses the Echo web framework for routing and middleware.

### Redis Database

User information and logs are stored in Redis, ensuring fast and efficient data retrieval.

### JWT Token Authentication

JWT tokens are used to authenticate users. After signing up, users receive a token upon logging in, which they must include in subsequent requests.

### Middleware

Middleware functions are used to:
- Authenticate users using JWT tokens.
- Restrict access to certain routes based on user roles (e.g., only admins can access specific routes).

### Routes

1. **Signup**: `/signup`
   
   ```bash
   curl -X POST http://localhost:3000/signup -d '{"name":"Alice","email":"alice@example.com","password":"password123","usertype":"Applicant"}' -H "Content-Type: application/json"
   ```
2. **Login**: `/login`

   ```bash
   curl -X POST http://localhost:3000/login -d '{"email":"alice@example.com","password":"password123"}' -H "Content-Type: application/json"
   ```
3. **Post a Job**: `/admin/job`

   ```bash
   curl -X POST http://localhost:3000/admin/job -H "Authorization: Bearer your_jwt_token_here" -H "Content-Type: application/json" -d '{"title": "Go Developer", "description": "Develop go applications"}'
   ```
4. **Upload Resume**: `/uploadResume`

   ```bash
   curl -X POST -H "Authorization: Bearer your_jwt_token_here" localhost:3000/uploadResume
   ```

### Code Structure

- `main.go`: Entry point of the application.
- `application/app.go`: Creating and starting new app instance.
- `application/routes.go`: Define routes and middleware.
- `handler/handler.go`: Contains route handlers.
- `model/model.go`: Defines data models (e.g., User struct).
