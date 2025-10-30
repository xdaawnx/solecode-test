package user

import (
	"fmt"
	"solecode/src/entities"
	"strings"
	"time"
)

func (uc *userUseCase) CreateUser(name, email string) (*entities.User, error) {
	// Check if email already exists
	existingUser, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	user := &entities.User{
		Name:  strings.TrimSpace(name),
		Email: strings.ToLower(strings.TrimSpace(email)),
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUseCase) GetUser(id int64) (*entities.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("user:%d", id)
	var user entities.User
	err := uc.cache.GetJSON(cacheKey, &user)
	if err == nil && user.ID != 0 {
		fmt.Printf("get user id %d from redis", id)
		return &user, nil
	}

	// If not in cache, get from database
	userPtr, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Cache the user data for 1 hour
	uc.cache.SetJSON(cacheKey, userPtr, time.Hour)

	return userPtr, nil
}

func (uc *userUseCase) UpdateUser(id int64, name, email string) (*entities.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get existing user
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if it's already taken by another user
	if user.Email != strings.ToLower(email) {
		existingUser, err := uc.userRepo.GetByEmail(email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
	}

	user.Name = strings.TrimSpace(name)
	user.Email = strings.ToLower(strings.TrimSpace(email))

	if err := uc.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	uc.cache.Delete(cacheKey)

	return user, nil
}

func (uc *userUseCase) DeleteUser(id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	if err := uc.userRepo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%d", id)
	uc.cache.Delete(cacheKey)

	return nil
}
