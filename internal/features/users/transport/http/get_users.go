package users_transport_http

import (
	"fmt"
	"net/http"

	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	core_http_utils "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/request"
	core_http_response "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/response"
)

type GetUsersResponse []UserDTOResponse

// GetUsers 		godoc
// @Summary 		Получить пользователей
// @Description 	Получить список пользователей с опциональной пагинацией
// @Tags 			users
// @Produce 		json
// @Param			limit query int false "Размер списка с пользователями"
// @Param			offset query int false "Отступ от начала списка с пользователями"
// @Success 		200 {object} GetUsersResponse "Успешно полученный пользователь"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/users [get]
func (h *UsersHTTPHandler) GetUsers(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	limit, offset, err := getLimitOffsetQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get \"limit\"/\"offset\" query param")
		return
	}

	usersDomain, err := h.usersService.GetUsers(ctx, limit, offset)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get users")
		return
	}

	response := GetUsersResponse(usersDTOFromDomains(usersDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func getLimitOffsetQueryParams(r *http.Request) (*int, *int, error) {
	var (
		limitQueryParamKey  = "limit"
		offsetQueryParamKey = "offset"
	)
	limit, err := core_http_utils.GetIntQueryParam(r, limitQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get '%s' query param: %w", limitQueryParamKey, err)
	}

	offset, err := core_http_utils.GetIntQueryParam(r, offsetQueryParamKey)
	if err != nil {
		return nil, nil, fmt.Errorf("get '%s' query param: %w", offsetQueryParamKey, err)
	}

	return limit, offset, nil
}
