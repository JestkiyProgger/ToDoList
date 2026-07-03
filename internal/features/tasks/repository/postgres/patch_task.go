package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
	core_postgres_pool "github.com/JestkiyProgger/ToDoList/internal/core/repository/postgres/pool"
)

func (r *TasksRepository) PatchTask(ctx context.Context, taskId int, task domain.Task) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE ToDoList.tasks
	SET
	    title=$1,
	    description=$2,
	    completed=$3,
	    completed_at=$4,
	    version=version+1
	
	WHERE id=$5 AND version=$6
	
	RETURNING id, version, title, description, completed, created_at, completed_at, user_id;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CompletedAt,
		taskId,
		task.Version,
	)

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
			return domain.Task{}, fmt.Errorf("task with id='%d' concurrently accessed: %w",
				taskId, core_errors.ErrConflict)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)

	return taskDomain, nil
}
