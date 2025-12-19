# Student Achievement Reporting System - Backend

Complete backend for student achievement reporting system built with Go Fiber, PostgreSQL, and MongoDB.

## Project Structure

```
├── main.go                 # Application entry point
├── config.go              # Database configurations
├── migrations.go           # Database migrations
├── models/
│   ├── user.go            # User and Auth models
│   ├── achievement.go      # Achievement models
│   └── student.go          # Student and Lecturer models
├── repository/            # Data access layer
│   ├── user_repository.go
│   ├── achievement_repository.go
│   ├── student_repository.go
│   └── lecturer_repository.go
├── service/               # Business logic layer
│   ├── auth_service.go
│   ├── achievement_service.go
│   ├── user_service.go
│   └── student_service.go
├── middleware/
│   └── auth.go            # Authentication and RBAC middleware
├── routes/                # API route handlers
│   ├── auth.go
│   ├── user.go
│   ├── achievement.go
│   ├── student.go
│   ├── lecturer.go
│   └── report.go
└── go.mod                 # Go module file
```

## Setup & Installation

1. **Prerequisites**
   - Go 1.21+
   - PostgreSQL 14+
   - MongoDB 5+

2. **Environment Setup**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

4. **Run Application**
   ```bash
   go run main.go
   ```

## Database Setup

### PostgreSQL
- Create database: `createdb achievement_db`
- Migrations run automatically on startup

### MongoDB
- Default: `mongodb://localhost:27017`
- Database: `achievement_db`
- Collections created automatically

## API Documentation

See Swagger documentation at `/swagger/index.html` after running the application.

## Architecture

### Clean Architecture Pattern
- **Models**: Data structures and interfaces
- **Repository**: Data access layer abstraction
- **Service**: Business logic and use cases
- **Routes**: HTTP handlers and request processing
- **Middleware**: Cross-cutting concerns (auth, RBAC)

### Database Design
- **PostgreSQL**: Relational data (users, roles, references)
- **MongoDB**: Dynamic achievement data

## Key Features

- ✅ JWT Authentication
- ✅ Role-Based Access Control (RBAC)
- ✅ Achievement Management (CRUD + Workflow)
- ✅ Multi-role support (Admin, Student, Lecturer)
- ✅ RESTful API Design
- ✅ Clean Architecture
- ✅ Comprehensive Error Handling

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/profile` - Get current user profile

### Users (Admin)
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Achievements
- `GET /api/v1/achievements` - List achievements
- `GET /api/v1/achievements/:id` - Get achievement
- `POST /api/v1/achievements` - Create achievement
- `POST /api/v1/achievements/:id/submit` - Submit for verification
- `POST /api/v1/achievements/:id/verify` - Verify achievement
- `POST /api/v1/achievements/:id/reject` - Reject achievement
- `DELETE /api/v1/achievements/:id` - Delete achievement

### Students
- `GET /api/v1/students` - List students
- `GET /api/v1/students/:id` - Get student
- `GET /api/v1/students/:id/achievements` - Student achievements
- `PUT /api/v1/students/:id/advisor` - Set advisor

### Lecturers
- `GET /api/v1/lecturers` - List lecturers
- `GET /api/v1/lecturers/:id` - Get lecturer
- `GET /api/v1/lecturers/:id/advisees` - Get advisees

### Reports
- `GET /api/v1/reports/statistics` - Achievement statistics
- `GET /api/v1/reports/student/:id` - Student report
