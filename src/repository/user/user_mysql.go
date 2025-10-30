package user

import (
	"database/sql"
	"fmt"
	"time"

	entities "solecode/src/entities"
)

func (r *userRepository) Create(user *entities.User) error {
	query := `
		INSERT INTO users (name, email, created_at, updated_at) 
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(query, user.Name, user.Email, now, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

func (r *userRepository) GetByID(id int64) (*entities.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at, deleted_at 
		FROM users 
		WHERE id = ? AND deleted_at IS NULL
	`

	user := &entities.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at, deleted_at 
		FROM users 
		WHERE email = ? AND deleted_at IS NULL
	`

	user := &entities.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *userRepository) Update(user *entities.User) error {
	query := `
		UPDATE users 
		SET name = ?, email = ?, updated_at = ? 
		WHERE id = ? AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()
	result, err := r.db.Exec(query, user.Name, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(id int64) error {
	query := `UPDATE users SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
