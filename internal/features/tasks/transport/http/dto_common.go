package tasks_transport_http

import (
	"time"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
)

type TaskDTOResponse struct {
	ID           int        `json:"id" example:"1"`
	Version      int        `json:"version" example:"1"`
	Title        string     `json:"title" example:"Сделать задачу"`
	Description  *string    `json:"description" example:"С помощью языка go"`
	Completed    bool       `json:"completed" example:"false"`
	CreatedAt    time.Time  `json:"create_at" example:"2026-01-02T10:25:00Z"`
	CompletedAt  *time.Time `json:"completed_at" example:"null"`
	AuthorUserID int        `json:"user_id" example:"1"`
}

func taskDTOFromDomain(taskDomain domain.Task) TaskDTOResponse {
	return TaskDTOResponse{
		ID:           taskDomain.ID,
		Version:      taskDomain.Version,
		Title:        taskDomain.Title,
		Description:  taskDomain.Description,
		Completed:    taskDomain.Completed,
		CreatedAt:    taskDomain.CreatedAt,
		CompletedAt:  taskDomain.CompletedAt,
		AuthorUserID: taskDomain.UserID,
	}
}

func taskDTOsFromDomains(tasksDomain []domain.Task) []TaskDTOResponse {
	dtos := make([]TaskDTOResponse, len(tasksDomain))
	for i, task := range tasksDomain {
		dtos[i] = taskDTOFromDomain(task)
	}

	return dtos
}
