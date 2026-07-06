package tasks_transport_http

import (
	"net/http"

	core_logger "github.com/JestkiyProgger/ToDoList/internal/core/logger"
	core_http_request "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/request"
	core_http_response "github.com/JestkiyProgger/ToDoList/internal/core/transport/http/response"
)

type GetTaskResponse TaskDTOResponse

// GetTask 			godoc
// @Summary 		Получить задачу
// @Description 	Получить конкретную задачу по ее ID
// @Tags 			tasks
// @Produce 		json
// @Param			id path int true "id получаемой задачи"
// @Success 		200 {object} GetTaskResponse "Успешно полученная задача"
// @Failure 		400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 		404 {object} core_http_response.ErrorResponse "Task not found"
// @Failure 		500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 			/tasks/{id} [get]
func (h *TaskHTTPHandler) GetTask(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	taskID, err := core_http_request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get taskID path value")
		return
	}

	taskDomain, err := h.tasksService.GetTask(ctx, *taskID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get task")
		return
	}

	response := GetTaskResponse(taskDTOFromDomain(taskDomain))

	responseHandler.JSONResponse(response, http.StatusOK)
}
