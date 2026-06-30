package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
)

func GetIntPathValue(r *http.Request, key string) (*int, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return nil, fmt.Errorf("no key='%s' in path values: %w", key, core_errors.ErrInvalidArgument)
	}

	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return nil, fmt.Errorf("path value='%s' by key='%s' no valid integer: %v, %w",
			pathValue, key, err, core_errors.ErrInvalidArgument)
	}

	return &val, nil
}
