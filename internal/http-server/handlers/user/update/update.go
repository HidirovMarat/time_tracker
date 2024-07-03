package update

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
	Id             int    `json:"id"`
	PassportNumber int    `json:"passportNumber"`
	PassportSerie  int    `json:"passportSerie"`
	Address        string `json:"address"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	Patronymic     string `json:"patronymic"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

type UserUpdate interface {
	UpdateUser(ctx context.Context, id int, passportSerie, passportNumber int, surname, name, patronymic string, address string) error
}

// @Summary Изменить user
// @Description patch user by request date
// @ID patch-user-by-user-field
// @Accept  json
// @Produce  text/plain
// @Success 200 "ok"
// @Failure 400 {string} string "empty body!!"
// @Failure 404 {string} string "Can not find ID"
// @Router /user [patch]
func New(context context.Context, log *slog.Logger, userUpdate UserUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.create.New"

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

		err = userUpdate.UpdateUser(context, req.Id, req.PassportSerie, req.PassportNumber, req.Surname, req.Name, req.Patronymic, req.Address)
		if err != nil {
			log.Error("failed to update user", err)
			http.Error(w, "Can not find ID", http.StatusInternalServerError)
			return
		}

		log.Info("user update", slog.Any("new field for user", req))

		w.WriteHeader(http.StatusOK)
	}
}
