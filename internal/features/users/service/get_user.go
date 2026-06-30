package users_service

import (
	"context"
	"fmt"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
)

func (us *UsersService) GetUser(ctx context.Context, id int) (domain.User, error) {
	user, err := us.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}

	return user, nil
}
