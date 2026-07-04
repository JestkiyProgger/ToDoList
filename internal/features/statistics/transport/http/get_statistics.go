package statistics_transport_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	core_http_request "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/request"
	core_http_response "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/response"
)

type GetStatisticsResponse struct {
	TaskCreated                int      `json:"task_created"`
	TaskCompleted              int      `json:"task_completed"`
	TaskCompletedRate          *float64 `json:"task_completed_rate"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time"`
}

func (h *StatisticsHTTPHandler) GetStatistics(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	userID, from, to, err := getUserIDFromToQueryParams(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get 'user_id'/'from'/'to' query params")
		return
	}

	statistics, err := h.statisticsService.GetStatistics(ctx, userID, from, to)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get statistics")
		return
	}

	response := toDTOFromDomain(statistics)

	responseHandler.JSONResponse(response, http.StatusOK)
}

func toDTOFromDomain(statistics domain.Statistics) GetStatisticsResponse {
	var avgTime *string
	if statistics.TasksAverageCompletionTime != nil {
		duration := statistics.TasksAverageCompletionTime.String()
		avgTime = &duration
	}

	return GetStatisticsResponse{
		TaskCreated:                statistics.TasksCreated,
		TaskCompleted:              statistics.TasksCompleted,
		TaskCompletedRate:          statistics.TasksCompletedRate,
		TasksAverageCompletionTime: avgTime,
	}
}

func getUserIDFromToQueryParams(r *http.Request) (*int, *time.Time, *time.Time, error) {
	var (
		userIDQueryParamKey = "user_id"
		fromQueryParamKey   = "from"
		toQueryParamKey     = "to"
	)

	userID, err := core_http_request.GetIntQueryParam(r, userIDQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get 'user_id' query param: %w", err)
	}

	from, err := core_http_request.GetDateQueryParam(r, fromQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get 'from' query param: %w", err)
	}

	to, err := core_http_request.GetDateQueryParam(r, toQueryParamKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get 'to' query param: %w", err)
	}

	return userID, from, to, nil
}
