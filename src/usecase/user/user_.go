package user

import (
	cachePkg "solecode/pkg/cache"
	"solecode/src/entities"
	userRepository "solecode/src/repository/user"
)

//go:generate mockery --name UserUseCaseItf --output mocks --filename userusecase_mock.go --outpkg mocks
type UserUseCaseItf interface {
	CreateUser(name, email string) (*entities.User, error)
	GetUser(id int64) (*entities.User, error)
	UpdateUser(id int64, name, email string) (*entities.User, error)
	DeleteUser(id int64) error
}

type userUseCase struct {
	userRepo userRepository.UserRepositoryItf
	cache    cachePkg.CacheItf
}

func NewUserUseCase(userRepo userRepository.UserRepositoryItf, cache cachePkg.CacheItf) UserUseCaseItf {
	return &userUseCase{
		userRepo: userRepo,
		cache:    cache,
	}
}
