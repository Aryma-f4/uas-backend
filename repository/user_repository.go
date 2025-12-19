package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Aryma-f4/uas-backend/models"
	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FullName, user.RoleID, user.IsActive,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at
		FROM users u
		WHERE u.username = $1 AND u.is_active = true
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at
		FROM users u
		WHERE u.id = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
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

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, role_id = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		user.Username, user.Email, user.FullName, user.RoleID, user.IsActive, user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE users SET is_active = false WHERE id = $1"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       u.role_id, u.is_active, u.created_at, u.updated_at
		FROM users u
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.FullName, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	query := `
		SELECT id, name, description, created_at
		FROM roles
		WHERE id = $1
	`

	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return role, nil
}
