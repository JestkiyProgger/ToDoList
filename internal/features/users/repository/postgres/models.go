package users_postgres_repository

import (
	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
)

type UserModel struct {
	ID          int
	Version     int
	Name        string
	PhoneNumber *string
}

func userDomainsFromModels(users []UserModel) []domain.User {
	userDomains := make([]domain.User, len(users))
	for i, user := range users {
		userDomains[i] = domain.NewUser(user.ID, user.Version, user.Name, user.PhoneNumber)
	}

	return userDomains
}
