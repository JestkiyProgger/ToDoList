package users_transport_http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id int) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) PatchUser(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error) {
	args := m.Called(ctx, id, patch)
	return args.Get(0).(domain.User), args.Error(1)
}

func returnPointer[T any](val T) *T {
	return &val
}

func createTargetGetUsers(limit *int, offset *int) string {
	query := url.Values{}

	if limit != nil {
		query.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		query.Set("offset", strconv.Itoa(*offset))
	}

	if len(query) == 0 {
		return "/users"
	}

	return "/users?" + query.Encode()
}

func newTestLogger() *core_logger.Logger {
	return &core_logger.Logger{
		Logger: zap.NewNop(),
	}
}

func addLoggerToContext(ctx context.Context) context.Context {
	return core_logger.ToContext(ctx, newTestLogger())
}

func setupTestRouter(handler *UsersHTTPHandler) *http.ServeMux {
	mux := http.NewServeMux()
	for _, route := range handler.Routes() {
		pattern := route.Method + " " + route.Path
		mux.HandleFunc(pattern, route.Handler)
	}
	return mux
}

func TestPatchUser_Success(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		requestBody  string // Изменили на string
		returnedUser domain.User
	}{
		{
			name:        "update both name and phone",
			userID:      1,
			requestBody: `{"name": "New Name", "phone_number": "+79998887766"}`,
			returnedUser: domain.User{
				ID:          1,
				Version:     2,
				Name:        "New Name",
				PhoneNumber: returnPointer("+79998887766"),
			},
		},
		{
			name:        "update only name",
			userID:      1,
			requestBody: `{"name": "Updated Name"}`,
			returnedUser: domain.User{
				ID:          1,
				Version:     2,
				Name:        "Updated Name",
				PhoneNumber: returnPointer("+79879879876"),
			},
		},
		{
			name:        "update only phone",
			userID:      1,
			requestBody: `{"phone_number": "+71112223344"}`,
			returnedUser: domain.User{
				ID:          1,
				Version:     2,
				Name:        "Ivan Ivanov",
				PhoneNumber: returnPointer("+71112223344"),
			},
		},
		{
			name:        "set phone to null (delete phone)",
			userID:      1,
			requestBody: `{"phone_number": null}`,
			returnedUser: domain.User{
				ID:          1,
				Version:     2,
				Name:        "Ivan Ivanov",
				PhoneNumber: nil,
			},
		},
		{
			name:        "empty request body (no changes)",
			userID:      1,
			requestBody: `{}`,
			returnedUser: domain.User{
				ID:          1,
				Version:     1,
				Name:        "Ivan Ivanov",
				PhoneNumber: returnPointer("+79879879876"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			mockService.
				On("PatchUser", mock.Anything, tt.userID, mock.AnythingOfType("domain.UserPatch")).
				Return(tt.returnedUser, nil)

			mux := setupTestRouter(handler)
			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d", tt.userID), bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			mockService.AssertExpectations(t)

			var response UserDTOResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.returnedUser.ID, response.ID)
			assert.Equal(t, tt.returnedUser.Version, response.Version)
			assert.Equal(t, tt.returnedUser.Name, response.Name)

			if tt.returnedUser.PhoneNumber != nil {
				require.NotNil(t, response.PhoneNumber)
				assert.Equal(t, *tt.returnedUser.PhoneNumber, *response.PhoneNumber)
			} else {
				assert.Nil(t, response.PhoneNumber)
			}
		})
	}
}

func TestPatchUser_ValidationError(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
	}{
		{
			name:        "name is null (forbidden)",
			requestBody: `{"name": null}`,
		},
		{
			name:        "name too short (< 3 symbols)",
			requestBody: `{"name": "Va"}`,
		},
		{
			name:        "name too long (> 100 symbols)",
			requestBody: `{"name": "` + strings.Repeat("a", 101) + `"}`,
		},
		{
			name:        "phone doesn't start with '+'",
			requestBody: `{"phone_number": "79879879876"}`,
		},
		{
			name:        "phone too short (< 10 symbols)",
			requestBody: `{"phone_number": "+7987"}`,
		},
		{
			name:        "phone too long (> 15 symbols)",
			requestBody: `{"phone_number": "+1234567890123456"}`,
		},
		{
			name:        "invalid JSON",
			requestBody: `{invalid json}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			mux := setupTestRouter(handler)
			req := httptest.NewRequest(http.MethodPatch, "/users/1", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertNotCalled(t, "PatchUser")
		})
	}
}

func TestPatchUser_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"letter instead of number", "/users/abc"},
		{"special characters", "/users/!@#"},
		{"float number", "/users/1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			mux := setupTestRouter(handler)
			req := httptest.NewRequest(http.MethodPatch, tt.path, bytes.NewReader([]byte(`{"name": "Test"}`)))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertNotCalled(t, "PatchUser")
		})
	}
}

func TestPatchUser_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.
		On("PatchUser", mock.Anything, 999, mock.AnythingOfType("domain.UserPatch")).
		Return(domain.User{}, core_errors.ErrNotFound)

	requestBody := `{"name": "New Name"}`

	mux := setupTestRouter(handler)
	req := httptest.NewRequest(http.MethodPatch, "/users/999", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestPatchUser_Conflict(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.
		On("PatchUser", mock.Anything, 1, mock.AnythingOfType("domain.UserPatch")).
		Return(domain.User{}, core_errors.ErrConflict)

	requestBody := `{"name": "New Name"}`

	mux := setupTestRouter(handler)
	req := httptest.NewRequest(http.MethodPatch, "/users/1", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockService.AssertExpectations(t)
}

func TestPatchUser_ServerError(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.
		On("PatchUser", mock.Anything, 1, mock.AnythingOfType("domain.UserPatch")).
		Return(domain.User{}, errors.New("database connection failed"))

	requestBody := `{"name": "New Name"}`

	mux := setupTestRouter(handler)
	req := httptest.NewRequest(http.MethodPatch, "/users/1", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUser_ServerError(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.
		On("DeleteUser", mock.Anything, mock.Anything).
		Return(errors.New("database error"))

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUser_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"letter instead of number", "/users/a"},
		{"special characters", "/users/!@#"},
		{"float number", "/users/1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			mux := setupTestRouter(handler)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertNotCalled(t, "DeleteUser")
		})
	}
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.
		On("DeleteUser", mock.Anything, 1).
		Return(core_errors.ErrNotFound)

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.On("DeleteUser", mock.Anything, 1).Return(nil)

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	mockService.AssertExpectations(t)
}

func TestGetUser_UserNotFound(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.On("GetUser", mock.Anything, 999).Return(domain.User{}, core_errors.ErrNotFound)

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestGetUser_ServerError(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.On("GetUser", mock.Anything, 1).Return(domain.User{}, errors.New("database error"))

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestGetUser_InvalidId(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/a", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetUser")
}

func TestGetUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	phoneNumber := "+79879879876"
	user := domain.User{
		ID:          1,
		Version:     1,
		Name:        "Ivan Ivanov",
		PhoneNumber: &phoneNumber,
	}

	mockService.
		On("GetUser", mock.Anything, mock.Anything).
		Once().
		Return(user, nil)

	mux := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)

	var response UserDTOResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, 1, response.Version)
	assert.Equal(t, "Ivan Ivanov", response.Name)
	require.NotNil(t, response.PhoneNumber)
	assert.Equal(t, "+79879879876", *response.PhoneNumber)
}

func TestGetUsers_ServerError(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	mockService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).
		Return([]domain.User{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	handler.GetUsers(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}

func TestGetUsers_InvalidOffset(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	req := httptest.NewRequest("GetUsers", "/users?offset=v", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	handler.GetUsers(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetUsers")
}

func TestGetUsers_InvalidLimit(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	req := httptest.NewRequest("GetUsers", "/users?limit=a", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	handler.GetUsers(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetUsers")
}

func TestGetUsers_ParsesLimitAndOffset(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	limit := 1
	offset := 1

	mockService.
		On("GetUsers", mock.Anything, &limit, &offset).
		Once().
		Return([]domain.User{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users?limit=1&offset=1", nil)
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	handler.GetUsers(rr, req)

	mockService.AssertExpectations(t)
}

func TestGetUsers_Success(t *testing.T) {
	tests := []struct {
		name          string
		expectedUsers []domain.User
		limit         *int
		offset        *int
	}{
		{
			name: "without a query params",
			expectedUsers: []domain.User{
				{
					ID:          1,
					Version:     1,
					Name:        "Ivan Ivanov",
					PhoneNumber: returnPointer("+79879879876"),
				},
				{
					ID:          2,
					Version:     1,
					Name:        "Petr Petrov",
					PhoneNumber: returnPointer("+79879879875"),
				},
			},
			limit:  nil,
			offset: nil,
		},
		{
			name: "with a 'limit' query params",
			expectedUsers: []domain.User{
				{
					ID:          1,
					Version:     1,
					Name:        "Ivan Ivanov",
					PhoneNumber: returnPointer("+79879879876"),
				},
			},
			limit:  returnPointer(1),
			offset: nil,
		},
		{
			name: "with a 'offset' query params",
			expectedUsers: []domain.User{
				{
					ID:          2,
					Version:     1,
					Name:        "Petr Petrov",
					PhoneNumber: returnPointer("+79879879876"),
				},
			},
			limit:  nil,
			offset: returnPointer(1),
		},
		{
			name:          "empty 'expectedUsers'",
			expectedUsers: []domain.User{},
			limit:         nil,
			offset:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			mockService.
				On("GetUsers", mock.Anything, tt.limit, tt.offset).
				Once().
				Return(tt.expectedUsers, nil)

			target := createTargetGetUsers(tt.limit, tt.offset)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()

			handler.GetUsers(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			mockService.AssertExpectations(t)

			var response []UserDTOResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			require.Len(t, response, len(tt.expectedUsers))

			for i := range len(tt.expectedUsers) {
				assert.Equal(t, tt.expectedUsers[i].ID, response[i].ID)
				assert.Equal(t, tt.expectedUsers[i].Version, response[i].Version)
				assert.Equal(t, tt.expectedUsers[i].Name, response[i].Name)
				require.NotNil(t, response[i].PhoneNumber)
				require.NotNil(t, tt.expectedUsers[i].PhoneNumber)
				assert.Equal(t, *tt.expectedUsers[i].PhoneNumber, *response[i].PhoneNumber)
			}
		})
	}
}

func TestCreateUser_Success(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	phoneNumber := "+79879879876"
	requestBody := CreateUserRequest{
		Name:        "Ivan Ivanov",
		PhoneNumber: &phoneNumber,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	expectedUser := domain.User{
		ID:          1,
		Version:     1,
		Name:        "Ivan Ivanov",
		PhoneNumber: &phoneNumber,
	}
	mockService.
		On("CreateUser", mock.Anything, mock.Anything).
		Once().
		Return(expectedUser, nil)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response CreateUserResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 1, response.ID)
	assert.Equal(t, 1, response.Version)
	assert.Equal(t, "Ivan Ivanov", response.Name)
	assert.Equal(t, "+79879879876", *response.PhoneNumber)

	mockService.AssertExpectations(t)
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(addLoggerToContext(req.Context()))
	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "CreateUser")
}

func TestCreateUser_ValidationError(t *testing.T) {
	phoneNumbers := []string{"79879879876", "+7", "+1234567890123456"}

	tests := []struct {
		name        string
		requestBody CreateUserRequest
	}{
		{
			name: "'Name' does not 3 symbols",
			requestBody: CreateUserRequest{
				Name: "Va",
			},
		},
		{
			name: "'PhoneNumbers' does not start with '+'",
			requestBody: CreateUserRequest{
				Name:        "Ivan",
				PhoneNumber: &phoneNumbers[0],
			},
		},
		{
			name: "length of 'PhoneNumbers' is less than 10 characters",
			requestBody: CreateUserRequest{
				Name:        "Ivan",
				PhoneNumber: &phoneNumbers[1],
			},
		},
		{
			name: "length of 'PhoneNumbers' exceeds 15 characters",
			requestBody: CreateUserRequest{
				Name:        "Ivan",
				PhoneNumber: &phoneNumbers[2],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			handler := NewUsersHTTPHandler(mockService)

			requestBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(addLoggerToContext(req.Context()))

			rr := httptest.NewRecorder()

			handler.CreateUser(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertNotCalled(t, "CreateUser")
		})
	}
}

func TestCreateUser_ServerError(t *testing.T) {
	mockService := new(MockUserService)
	handler := NewUsersHTTPHandler(mockService)

	requestBody := `{"name": "Ivan Ivanov"}`

	mockService.
		On("CreateUser", mock.Anything, mock.Anything).
		Once().
		Return(domain.User{}, errors.New("database error"))

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(addLoggerToContext(req.Context()))

	rr := httptest.NewRecorder()

	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}
