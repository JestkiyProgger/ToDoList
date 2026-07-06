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
	TaskCreated                int      `json:"task_created" example:"2"`
	TaskCompleted              int      `json:"task_completed" example:"1"`
	TaskCompletedRate          *float64 `json:"task_completed_rate" example:"50"`
	TasksAverageCompletionTime *string  `json:"tasks_average_completion_time" example:"1h30m"`
}

// GetStatistics 	godoc
// @Summary 		Получение статистики
// @Description 	Получение: Количества созданных задач, количество сделанных задач, процент выполненных задач, среднее время выполнения задачи
// @Tags 			statistics
// @Produce 		json
// @Param			from 	query string false "Размер списка задач"
// @Param			to   	query string false "Отступ от начала списка задач"
// @Param			user_id query int    false "Статистика по определенного пользователя"
// @Success 		200 {object} GetStatisticsResponse "Успешно полученная статистика"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/statistics [get]
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
