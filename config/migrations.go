package config

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RunMigrations(db *sql.DB) error {
	// Create tables
	if err := createTables(db); err != nil {
		return err
	}

	// Seed initial data
	if err := seedData(db); err != nil {
		return err
	}

	return nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		// Roles table
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Permissions table
		`CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			resource VARCHAR(50) NOT NULL,
			action VARCHAR(50) NOT NULL,
			description TEXT
		)`,

		// Role permissions table
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		)`,

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			role_id UUID REFERENCES roles(id),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Lecturers table
		`CREATE TABLE IF NOT EXISTS lecturers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			lecturer_id VARCHAR(20) UNIQUE NOT NULL,
			department VARCHAR(100),
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Students table
		`CREATE TABLE IF NOT EXISTS students (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			student_id VARCHAR(20) UNIQUE NOT NULL,
			program_study VARCHAR(100),
			academic_year VARCHAR(10),
			advisor_id UUID REFERENCES lecturers(id),
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Achievement references table
		`CREATE TABLE IF NOT EXISTS achievement_references (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			student_id UUID REFERENCES students(id) ON DELETE CASCADE,
			mongo_achievement_id VARCHAR(24) NOT NULL,
			status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'verified', 'rejected')),
			submitted_at TIMESTAMP,
			verified_at TIMESTAMP,
			verified_by UUID REFERENCES users(id),
			rejection_note TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,

		// Achievement status history table
		`CREATE TABLE IF NOT EXISTS achievement_status_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			achievement_ref_id UUID REFERENCES achievement_references(id) ON DELETE CASCADE,
			old_status VARCHAR(20),
			new_status VARCHAR(20) NOT NULL,
			changed_by UUID REFERENCES users(id),
			note TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role_id)`,
		`CREATE INDEX IF NOT EXISTS idx_students_advisor ON students(advisor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_achievement_refs_student ON achievement_references(student_id)`,
		`CREATE INDEX IF NOT EXISTS idx_achievement_refs_status ON achievement_references(status)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Migration error: %v", err)
			return err
		}
	}

	return nil
}

func seedData(db *sql.DB) error {
	// Check if roles already exist
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	// Seed roles
	roles := []struct {
		ID          uuid.UUID
		Name        string
		Description string
	}{
		{uuid.New(), "Admin", "Administrator with full system access"},
		{uuid.New(), "Mahasiswa", "Student who can report achievements"},
		{uuid.New(), "Dosen Wali", "Academic advisor who verifies achievements"},
	}

	for _, role := range roles {
		_, err := db.Exec(
			"INSERT INTO roles (id, name, description) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING",
			role.ID, role.Name, role.Description,
		)
		if err != nil {
			return err
		}
	}

	// Get role IDs
	var adminRoleID, mahasiswaRoleID, dosenRoleID uuid.UUID
	db.QueryRow("SELECT id FROM roles WHERE name = 'Admin'").Scan(&adminRoleID)
	db.QueryRow("SELECT id FROM roles WHERE name = 'Mahasiswa'").Scan(&mahasiswaRoleID)
	db.QueryRow("SELECT id FROM roles WHERE name = 'Dosen Wali'").Scan(&dosenRoleID)

	// Seed permissions
	permissions := []struct {
		Name        string
		Resource    string
		Action      string
		Description string
	}{
		{"achievement:create", "achievement", "create", "Create new achievement"},
		{"achievement:read", "achievement", "read", "Read achievement data"},
		{"achievement:update", "achievement", "update", "Update achievement data"},
		{"achievement:delete", "achievement", "delete", "Delete achievement"},
		{"achievement:verify", "achievement", "verify", "Verify achievement"},
		{"achievement:reject", "achievement", "reject", "Reject achievement"},
		{"user:create", "user", "create", "Create new user"},
		{"user:read", "user", "read", "Read user data"},
		{"user:update", "user", "update", "Update user data"},
		{"user:delete", "user", "delete", "Delete user"},
		{"user:manage", "user", "manage", "Full user management"},
		{"student:read", "student", "read", "Read student data"},
		{"student:manage", "student", "manage", "Manage student data"},
		{"lecturer:read", "lecturer", "read", "Read lecturer data"},
		{"lecturer:manage", "lecturer", "manage", "Manage lecturer data"},
		{"report:read", "report", "read", "Read reports"},
		{"report:all", "report", "all", "Access all reports"},
	}

	permissionIDs := make(map[string]uuid.UUID)
	for _, perm := range permissions {
		id := uuid.New()
		permissionIDs[perm.Name] = id
		_, err := db.Exec(
			"INSERT INTO permissions (id, name, resource, action, description) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (name) DO NOTHING",
			id, perm.Name, perm.Resource, perm.Action, perm.Description,
		)
		if err != nil {
			return err
		}
	}

	// Refresh permission IDs from database
	rows, _ := db.Query("SELECT id, name FROM permissions")
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name string
		rows.Scan(&id, &name)
		permissionIDs[name] = id
	}

	// Assign permissions to roles
	// Admin gets all permissions
	for _, permID := range permissionIDs {
		db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", adminRoleID, permID)
	}

	// Mahasiswa permissions
	mahasiswaPerms := []string{"achievement:create", "achievement:read", "achievement:update", "achievement:delete", "report:read"}
	for _, perm := range mahasiswaPerms {
		if permID, ok := permissionIDs[perm]; ok {
			db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", mahasiswaRoleID, permID)
		}
	}

	// Dosen Wali permissions
	dosenPerms := []string{"achievement:read", "achievement:verify", "achievement:reject", "student:read", "report:read"}
	for _, perm := range dosenPerms {
		if permID, ok := permissionIDs[perm]; ok {
			db.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", dosenRoleID, permID)
		}
	}

	// Create default admin user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminID := uuid.New()
	_, err := db.Exec(
		`INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (username) DO NOTHING`,
		adminID, "admin", "admin@university.ac.id", string(hashedPassword), "System Administrator", adminRoleID, true,
	)
	if err != nil {
		log.Printf("Error creating admin user: %v", err)
	}

	log.Println("Database seeded successfully")
	return nil
}
