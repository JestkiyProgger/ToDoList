package tasks_postgres_repository

import (
	"context"
	"fmt"

	core_errors "github.com/JestkiyProgger/ToDoList/internal/core/errors"
)

func (r *TasksRepository) DeleteTask(ctx context.Context, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM ToDoList.tasks
	WHERE id=$1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("task with id='%d': %w", taskID, core_errors.ErrNotFound)
	}

	return nil
}
