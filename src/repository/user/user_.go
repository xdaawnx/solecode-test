package user

import (
	"database/sql"
	"solecode/src/entities"
)

//go:generate mockery --name UserRepositoryItf --output mocks --filename userrepository_mock.go --outpkg mocks
type UserRepositoryItf interface {
	Create(user *entities.User) error
	GetByID(id int64) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	Update(user *entities.User) error
	Delete(id int64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepositoryItf {
	return &userRepository{db: db}
}
