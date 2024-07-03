package start

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

type TaskStart interface {
	BeginTask(ctx context.Context, id int, startTime time.Time) error
}

// @Summary Начать task time
// @Description начать отчет времени task, поле start_time
// @ID put-task-of-start_time
// @Accept  json
// @Produce  text/plain
// @Success 200 "ok"
// @Failure 400 {string} string "empty body"
// @Failure 404 {string} string "have't task"
// @Router /task/start [put]
func New(context context.Context, log *slog.Logger, taskStart TaskStart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.start.New"

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

		err = taskStart.BeginTask(context, req.Id, time.Now())

		if err != nil {
			log.Error("failed to start task", err)
			http.Error(w, "have't task", http.StatusInternalServerError)
			return
		}

		log.Info("start task", slog.Int("id", req.Id))

		w.WriteHeader(http.StatusOK)
	}
}
