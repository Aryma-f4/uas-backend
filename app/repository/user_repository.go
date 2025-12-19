package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, r.name as role_name, u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, r.name as role_name, u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, r.name as role_name, u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive,
	)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users SET username = $2, email = $3, full_name = $4, is_active = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.FullName, user.IsActive)
	return err
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `UPDATE users SET role_id = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, r.name as role_name, u.is_active, u.created_at, u.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
			&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (r *UserRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *UserRepository) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	role := &entity.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserRepository) GetRoles(ctx context.Context) ([]*entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		role := &entity.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *UserRepository) CheckPermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1 AND p.name = $2
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, roleID, permission).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*entity.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, u.email, s.program_study, s.academic_year, s.advisor_id, s.created_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.user_id = $1
	`
	student := &entity.Student{}
	var advisorID sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.FullName, &student.Email,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if advisorID.Valid {
		id, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &id
	}
	return student, nil
}

func (r *UserRepository) GetLecturerByUserID(ctx context.Context, userID uuid.UUID) (*entity.Lecturer, error) {
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, u.email, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE l.user_id = $1
	`
	lecturer := &entity.Lecturer{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.FullName, &lecturer.Email,
		&lecturer.Department, &lecturer.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("lecturer not found: %w", err)
	}
	return lecturer, nil
}
