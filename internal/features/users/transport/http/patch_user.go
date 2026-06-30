package users_transport_http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	core_http_request "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/request"
	core_http_response "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/response"
	core_http_types "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/types"
)

type PatchUserRequest struct {
	Name        core_http_types.Nullable[string] `json:"name"`
	PhoneNumber core_http_types.Nullable[string] `json:"phone_number"`
}

func (r *PatchUserRequest) Validate() error {
	if r.Name.Set {
		if r.Name.Value == nil {
			return fmt.Errorf("'Name' can't be NULL")
		}

		nameLen := len([]rune(*r.Name.Value))
		if nameLen < 3 || nameLen > 100 {
			return fmt.Errorf("'Name' must be between 3 and 100 symbols")
		}
	}

	if r.PhoneNumber.Set {
		if r.PhoneNumber.Value != nil {
			phoneNumberLen := len([]rune(*r.PhoneNumber.Value))
			if phoneNumberLen < 10 || phoneNumberLen > 15 {
				return fmt.Errorf("'Phone Number' must be between 10 and 15 symbols")
			}

			if !strings.HasPrefix(*r.PhoneNumber.Value, "+") {
				return fmt.Errorf("'Phone Number' must startswith '+' symbol")
			}
		}
	}

	return nil
}

type PatchUserResponse UserDTOResponse

func (h *UsersHTTPHandler) PatchUser(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userId, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")
		return
	}

	var request PatchUserRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate HTTP request")
		return
	}

	userPatch := userPatchFromRequest(request)

	userDomain, err := h.usersService.PatchUser(ctx, *userId, userPatch)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to patch user")
		return
	}

	response := PatchUserResponse(userDTOFromDomain(userDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}

func userPatchFromRequest(request PatchUserRequest) domain.UserPatch {
	return domain.NewUserPatch(request.Name.ToDomain(), request.PhoneNumber.ToDomain())
}
