package delete

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
	Id int `json:"passportNumber" validate:"required"`
}

type UserDelete interface {
	DeleteUser(ctx context.Context, id int) error
}

// @Summary Удалить user
// @Description удалить user по id 
// @ID delete-user-by-id
// @Accept  json
// @Produce text/plain
// @Success 200 "ok"
// @Failure 400 {string} string "empty body"
// @Failure 404 {string} string "have't user"
// @Router /user [delete]
func New(context context.Context, log *slog.Logger, userDelete UserDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.delete.New"

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

		err = userDelete.DeleteUser(context, req.Id)

		if err != nil {
			log.Error("failed to delete user", err)
			http.Error(w, "have't user", http.StatusInternalServerError)
			return
		}

		log.Info("user delete", slog.Int("id", req.Id))

		w.WriteHeader(http.StatusOK)
	}
}
