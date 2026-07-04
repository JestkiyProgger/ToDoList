package statistics_postgres_repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JestkiyProgger/ToDoList/internal/core/domain"
)

func (r *StatisticsRepository) GetTasks(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`
	SELECT id, version, title, description, completed, created_at, completed_at, user_id
	FROM ToDoList.tasks
	`)

	args := []any{}
	conditions := []string{}

	if userID != nil {
		args = append(args, userID)
		conditions = append(conditions, fmt.Sprintf("user_id=$%d", len(args)))
	}

	if from != nil {
		args = append(args, from)
		conditions = append(conditions, fmt.Sprintf("created_at>=$%d", len(args)))
	}

	if to != nil {
		args = append(args, to)
		conditions = append(conditions, fmt.Sprintf("created_at<$%d", len(args)))
	}

	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE " + strings.Join(conditions, " AND "))
	}

	queryBuilder.WriteString(" ORDER BY id ASC")
	rows, err := r.pool.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasksModel []TaskModel
	for rows.Next() {
		var taskModel TaskModel
		err = rows.Scan(
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
			return nil, fmt.Errorf("scan tasks: %w", err)
		}

		tasksModel = append(tasksModel, taskModel)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	taskDomains := taskDomainsFromModels(tasksModel)

	return taskDomains, nil
}
