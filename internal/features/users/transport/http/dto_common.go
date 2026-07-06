package users_transport_http

import "github.com/JestkiyProgger/ToDoList/internal/core/domain"

type UserDTOResponse struct {
	ID          int     `json:"id" example:"1"`
	Version     int     `json:"version" example:"1"`
	Name        string  `json:"name" example:"Ivan Ivanov"`
	PhoneNumber *string `json:"phone_number" example:"+79879879876"`
}

func userDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:          user.ID,
		Version:     user.Version,
		Name:        user.Name,
		PhoneNumber: user.PhoneNumber,
	}
}

func usersDTOFromDomains(users []domain.User) []UserDTOResponse {
	usersDTO := make([]UserDTOResponse, len(users))
	for i, user := range users {
		usersDTO[i] = userDTOFromDomain(user)
	}

	return usersDTO
}
