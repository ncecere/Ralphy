# Example Project PRD

This is an example Product Requirements Document (PRD) for use with Ralphy.
Tasks are defined using markdown checkbox syntax.

## Overview

Building a REST API for a todo application with user authentication.

## Tasks

### Phase 1: Setup

- [ ] Initialize Go module and project structure
- [ ] Set up configuration management with Viper
- [ ] Create database connection pool with PostgreSQL

### Phase 2: Authentication

- [ ] Implement user registration endpoint (POST /api/auth/register)
- [ ] Implement user login endpoint with JWT tokens (POST /api/auth/login)
- [ ] Add middleware for JWT token validation
- [ ] Implement password reset flow (POST /api/auth/reset)

### Phase 3: Core Features

- [ ] Create todo CRUD endpoints (GET/POST/PUT/DELETE /api/todos)
- [ ] Add pagination support for list endpoints
- [ ] Implement todo filtering by status and due date
- [ ] Add todo sharing between users

### Phase 4: Polish

- [ ] Add request validation with detailed error messages
- [ ] Implement rate limiting middleware
- [ ] Add OpenAPI/Swagger documentation
- [ ] Write integration tests for all endpoints

## Completed Tasks

- [x] Research authentication best practices
- [x] Design database schema
- [x] Set up CI/CD pipeline

## Notes

- Use chi router for HTTP handling
- Follow REST API best practices
- All endpoints should return JSON
- Use structured logging with slog
