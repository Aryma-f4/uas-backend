package main

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		createRolesTable,
		createPermissionsTable,
		createRolePermissionsTable,
		createUsersTable,
		createStudentsTable,
		createLecturersTable,
		createAchievementReferencesTable,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			log.Printf("Migration error: %v", err)
			return err
		}
	}

	// Insert default roles
	if err := insertDefaultRoles(db); err != nil {
		return err
	}

	// Insert default permissions
	if err := insertDefaultPermissions(db); err != nil {
		return err
	}

	return nil
}

const createRolesTable = `
CREATE TABLE IF NOT EXISTS roles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(50) UNIQUE NOT NULL,
	description TEXT,
	created_at TIMESTAMP DEFAULT NOW()
);
`

const createPermissionsTable = `
CREATE TABLE IF NOT EXISTS permissions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) UNIQUE NOT NULL,
	resource VARCHAR(50) NOT NULL,
	action VARCHAR(50) NOT NULL,
	description TEXT
);
`

const createRolePermissionsTable = `
CREATE TABLE IF NOT EXISTS role_permissions (
	role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
	permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
	PRIMARY KEY (role_id, permission_id)
);
`

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(50) UNIQUE NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	full_name VARCHAR(100) NOT NULL,
	role_id UUID REFERENCES roles(id),
	is_active BOOLEAN DEFAULT true,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);
`

const createStudentsTable = `
CREATE TABLE IF NOT EXISTS students (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	student_id VARCHAR(20) UNIQUE NOT NULL,
	program_study VARCHAR(100),
	academic_year VARCHAR(10),
	advisor_id UUID REFERENCES lecturers(id),
	created_at TIMESTAMP DEFAULT NOW()
);
`

const createLecturersTable = `
CREATE TABLE IF NOT EXISTS lecturers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	lecturer_id VARCHAR(20) UNIQUE NOT NULL,
	department VARCHAR(100),
	created_at TIMESTAMP DEFAULT NOW()
);
`

const createAchievementReferencesTable = `
CREATE TABLE IF NOT EXISTS achievement_references (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	student_id UUID REFERENCES students(id) ON DELETE CASCADE,
	mongo_achievement_id VARCHAR(24) NOT NULL,
	status VARCHAR(20) DEFAULT 'draft',
	submitted_at TIMESTAMP,
	verified_at TIMESTAMP,
	verified_by UUID REFERENCES users(id),
	rejection_note TEXT,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);
`

func insertDefaultRoles(db *sql.DB) error {
	roles := []struct {
		name        string
		description string
	}{
		{"Admin", "System administrator"},
		{"Mahasiswa", "Student"},
		{"Dosen Wali", "Academic advisor"},
	}

	for _, role := range roles {
		_, err := db.Exec(
			"INSERT INTO roles (name, description) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			role.name, role.description,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertDefaultPermissions(db *sql.DB) error {
	permissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"achievement:create", "achievement", "create", "Create achievement"},
		{"achievement:read", "achievement", "read", "Read achievement"},
		{"achievement:update", "achievement", "update", "Update achievement"},
		{"achievement:delete", "achievement", "delete", "Delete achievement"},
		{"achievement:verify", "achievement", "verify", "Verify achievement"},
		{"user:manage", "user", "manage", "Manage users"},
	}

	for _, perm := range permissions {
		_, err := db.Exec(
			"INSERT INTO permissions (name, resource, action, description) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			perm.name, perm.resource, perm.action, perm.description,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
