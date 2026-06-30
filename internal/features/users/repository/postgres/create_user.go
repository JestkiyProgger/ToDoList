package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `INSERT INTO todolist.users (name, phone_number) VALUES ($1, $2) RETURNING id, version, name, phone_number;`

	row := r.pool.QueryRow(ctx, query, user.Name, user.PhoneNumber)

	var userModel UserModel
	err := row.Scan(&userModel.ID, &userModel.Version, &userModel.Name, &userModel.PhoneNumber)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	userDomain := domain.NewUser(userModel.ID, userModel.Version, userModel.Name, userModel.PhoneNumber)

	return userDomain, nil
}
