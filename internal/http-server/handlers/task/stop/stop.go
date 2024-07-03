package stop

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Id int `json:"id" validate:"required"`
}

type TaskStop interface {
	StopTask(ctx context.Context, id int, endTime time.Time) error
}

// @Summary Остановить task time
// @Description остановить отчет времени task, поле stop_time
// @ID put-task-of-stop_time
// @Accept  json
// @Produce  text/plain
// @Success 200 "ok"
// @Failure 400 {string} string "empty body"
// @Failure 404 {string} string "have't task"
// @Router /task/stop [put]
func New(context context.Context, log *slog.Logger, taskStop TaskStop) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.stop.New"

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

		err = taskStop.StopTask(context, req.Id, time.Now())

		if err != nil {
			log.Error("failed to stop task", err)
			http.Error(w, "have't task", http.StatusInternalServerError)
			return
		}

		log.Info("stop task", slog.Int("id", req.Id))

		w.WriteHeader(http.StatusOK)
	}
}
