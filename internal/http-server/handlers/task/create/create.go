package create

import (
	"context"
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	UserId      int    `json:"user_id" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type Response struct {
	Id int `json:"id,omitempty"`
}

type TaskCreate interface {
	CreateTask(ctx context.Context, userId int, description string) (int, error)
}

// @Summary Создать task
// @Description создать task по user_id и description
// @ID create-task-by-user_id-description
// @Accept  json
// @Produce  json
// @Success 200 {int} id "ok"
// @Failure 400 {string} string "empty body"
// @Failure 404 {string} string "not save task"
// @Router /task [post]
func New(context context.Context, log *slog.Logger, taskCreate TaskCreate) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.task.create.New"

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

		id, err := taskCreate.CreateTask(context, req.UserId, req.Description)

		if err != nil {
			log.Error("failed to add task", err)
			http.Error(w, "not save task", http.StatusInternalServerError)
			return
		}

		log.Info("task added", slog.Int("id", id))

		responseOK(w, r, id)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Id: id,
	})
}
