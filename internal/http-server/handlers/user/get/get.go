package get

import (
	"context"
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"time_tracker/internal/storage/post"
)

type Request struct {
	Id             *int    `json:"id"`
	PassportNumber *int    `json:"passportNumber"`
	PassportSerie  *int    `json:"passportSerie"`
	Address        *string `json:"address"`
	Name           *string `json:"name"`
	Surname        *string `json:"surname"`
	Patronymic     *string `json:"patronymic"`
	Limit          *int    `json:"limit"`
	Offset         *int    `json:"offset"`
}

type Response struct {
	Users []post.User `json:"users,omitempty"`
}

type UserGet interface {
	GetUser(ctx context.Context, id *int, passportSerie, passportNumber *int, surname, name, patronymic *string, address *string, offset, limit *int) ([]post.User, error)
}

// @Summary Получить user
// @Description получить user,также фильтрация и пагинация
// @ID get-user-by-id
// @Accept  json
// @Produce  json
// @Success 200 {array} post.User "ok"
// @Failure 400 {string} string "request body is empty"
// @Failure 404 {string} string "User not found"
// @Router /user [get]
func New(context context.Context, log *slog.Logger, userGet UserGet) http.HandlerFunc {
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

		users, err := userGet.GetUser(context, req.Id, req.PassportSerie, req.PassportNumber, req.Surname, req.Name, req.Patronymic, req.Address, req.Offset, req.Limit)
		if err != nil {
			log.Error("failed to get user", err)
			http.Error(w, "error to DB", http.StatusInternalServerError)
			return
		}

		log.Info("user get", slog.Any("User If", req))

		responseOK(w, r, users)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, users []post.User) {
	render.JSON(w, r, Response{
		Users: users,
	})
}
