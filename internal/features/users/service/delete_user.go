package users_service

import (
	"context"
	"fmt"
)

func (us *UsersService) DeleteUser(ctx context.Context, id int) error {
	if err := us.usersRepository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
