package getUserTasks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"time_tracker/internal/storage/post"
)

type Request struct {
	UserId      int       `json:"user_id"`
	StartPeriod time.Time `json:"startPeriod"`
	EndPeriod   time.Time `json:"endPeriod"`
}

type Response struct {
	TaskTimes []post.TaskTime `json:"task_time,omitempty"`
}

type UserTaskTimeGet interface {
	GetUserTaskTime(ctx context.Context, user_id int, startPeriod, endPeriod time.Time) ([]post.TaskTime, error)
}

// @Summary Получить userTaskTime
// @Description получить userTaskTime по user_id и startPerio, endPeriod
// @ID get-user_task_time-by-user_id-startPeriod-endPeriod
// @Accept  json
// @Produce  json
// @Success 200 {array} post.TaskTime "ok"
// @Failure 400 {string} string "empty body"
// @Failure 404 {string} string "failed to get user_task_time"
// @Router //task/task-time [get]
func New(context context.Context, log *slog.Logger, userTaskTimeGet UserTaskTimeGet) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.taks.getTaskTime.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Error("failed to decode request body", err)
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		taskTimes, err := userTaskTimeGet.GetUserTaskTime(context, req.UserId, req.StartPeriod, req.EndPeriod)
		if err != nil {
			log.Error("failed to get user_task_time", err)
			http.Error(w, "error to DB", http.StatusInternalServerError)
			return
		}

		log.Info("userTaskTime get", slog.Any("user_id", req.UserId))

		responseOK(w, r, taskTimes)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, taskTimes []post.TaskTime) {
	render.JSON(w, r, Response{
		TaskTimes: taskTimes,
	})
}
