package tasks_service

import (
	"context"
	"fmt"
)

func (s *TasksService) DeleteTask(ctx context.Context, taskID int) error {
	if err := s.tasksRepository.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("delete task from repository: %w", err)
	}

	return nil
}
