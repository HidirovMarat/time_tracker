package post

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type TaskTime struct {
	TaskID  int `json:"task_id"`
	Hours   float64 `json:"hours"`
	Minutes float64 `json:"minutes"`
}

func (pg *postgres) CreateTask(ctx context.Context, userId int, description string) (int, error) {
	query := `
	INSERT INTO tasks (user_id, description) 
	VALUES (@user_id, @description) RETURNING id`

	args := pgx.NamedArgs{
		"user_id":     userId,
		"description": description,
	}

	result := pg.db.QueryRow(ctx, query, args)

	var id int
	err := result.Scan(&id)

	if err != nil {
		return -1, fmt.Errorf("unable to insert row: %w", err)
	}

	return id, nil
}

func (pg *postgres) BeginTask(ctx context.Context, id int, startTime time.Time) error {
	query := `
	UPDATE tasks SET start_time = @start_time WHERE id = @id
	`

	args := pgx.NamedArgs{
		"start_time": startTime,
		"id":       id,
	}

	results, err := pg.db.Exec(ctx, query, args)

	if results.RowsAffected() == 0 {
		return errors.New("not id delet")
	}

	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (pg *postgres) StopTask(ctx context.Context, id int, endTime time.Time) error {
	query := `
	UPDATE tasks SET end_time = @end_time WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id":       id,
		"end_time": endTime,
	}

	results, err := pg.db.Exec(ctx, query, args)

	if results.RowsAffected() == 0 {
		return errors.New("not id delet")
	}

	if err != nil {
		return fmt.Errorf("unable to insert row: %w", err)
	}

	return nil
}

func (pg *postgres) GetUserTaskTime(ctx context.Context, user_id int, startPeriod, endPeriod time.Time) ([]TaskTime, error) {
	query := `
    SELECT tasks.id as task_id, 
	EXTRACT(EPOCH FROM (end_time - start_time)) / 3600 AS hours, 
	(EXTRACT(EPOCH FROM (end_time - start_time)) % 3600) / 60 AS minutes
    FROM tasks
	join users on tasks.user_id = users.id
	WHERE user_id = @user_id AND @start_period < start_time AND end_time < @end_period
	ORDER BY hours, minutes DESC
	`
	args := pgx.NamedArgs{
		"user_id":      user_id,
		"start_period": startPeriod,
		"end_period":   endPeriod,
	}

	rows, err := pg.db.Query(ctx, query, args)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[TaskTime])

	if err != nil {
		return nil, err
	}

	return result, err
}
