package usecases

import (
	"solecode/pkg/cache"
	repo "solecode/src/repository"
	userUC "solecode/src/usecase/user"
)

type UseCases struct {
	User userUC.UserUseCaseItf
}

func InitUsecase(
	repo repo.Repository,
	cache cache.CacheItf,
) *UseCases {
	// Initialize user use case
	userUseCase := userUC.NewUserUseCase(
		repo.User,
		cache,
	)

	return &UseCases{
		User: userUseCase,
	}
}
