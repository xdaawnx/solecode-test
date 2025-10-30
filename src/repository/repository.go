package repository

import (
	"database/sql"

	userRepo "solecode/src/repository/user"
)

type Repository struct {
	User userRepo.UserRepositoryItf
}

func InitRepository(db *sql.DB) *Repository {
	return &Repository{
		User: userRepo.NewUserRepository(db),
	}
}
