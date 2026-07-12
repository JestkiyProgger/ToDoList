package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
	core_postgres_pool "github.com/JestkiyProgger/ToDoList/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) GetTask(ctx context.Context, taskID int) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, title, description, completed, created_at, completed_at, user_id
	FROM ToDoList.tasks
	WHERE id=$1;
	`

	row := r.pool.QueryRow(ctx, query, taskID)

	var taskModel TaskModel
	err := row.Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.UserId,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf("task with id='%d': %w", taskID, core_errors.ErrNotFound)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)

	return taskDomain, nil
}
