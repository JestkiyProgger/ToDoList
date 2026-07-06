package users_transport_http

import (
	"net/http"

	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	core_http_utils "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/request"
	core_http_response "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/response"
)

type GetUserResponse UserDTOResponse

// GetUser 			godoc
// @Summary 		Получить пользователя
// @Description 	Получить конкретного пользователя по его ID
// @Tags 			users
// @Produce 		json
// @Param			id path int true "id получаемого пользователя"
// @Success 		200 {object} GetUserResponse "Успешно полученный пользователь"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		404 {object} core_http_response.ErrorResponse "User not found"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/users/{id} [get]
func (h *UsersHTTPHandler) GetUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userId, err := core_http_utils.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")
		return
	}

	user, err := h.usersService.GetUser(ctx, *userId)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user")
		return
	}

	response := GetUserResponse(userDTOFromDomain(user))

	responseHandler.JSONResponse(response, http.StatusOK)
}
