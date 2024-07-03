package create

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"time_tracker/internal/request/info"
)

type Request struct {
	Passport string `json:"passportNumber" validate:"required"`
}

type Response struct {
	Id int `json:"id,omitempty"`
}

type UserCreate interface {
	CreateUser(ctx context.Context, passportSerie, passportNumber int, surname, name, patronymic string, address string) (int, error)
}

type UserInfoGet interface {
	GetUserInfo(passportSerie int, passportNumber int, baseURL string) (*info.UserInfoResponse, error)
}

// @Summary Создать user
// @Description создать user по паспорту и получить данные через другой сервис 
// @ID create-user-by-passport
// @Accept  json
// @Produce  json
// @Success 200 {int} id "ok"
// @Failure 400 {string} string "not correct passport"
// @Failure 404 {string} string "not save user"
// @Router /user [post]
func New(context context.Context, log *slog.Logger, userCreate UserCreate, userInfoGet UserInfoGet, baseURL string) http.HandlerFunc {
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

		passport := strings.Fields(req.Passport)

		if len(passport) != 2 {
			log.Error("failed to decode json passport", err)
			http.Error(w, "not correct passport", http.StatusBadRequest)
			return
		}

		passportSerie, err := strconv.Atoi(passport[0])
		if err != nil {
			log.Error("failed to decode json passport", err)
			http.Error(w, "not correct passport", http.StatusBadRequest)
			return
		}

		passportNumber, err := strconv.Atoi(passport[1])
		if err != nil {
			log.Error("failed to decode json passport", err)
			http.Error(w, "not correct passport", http.StatusBadRequest)
			return
		}

		userInfoRes, err := userInfoGet.GetUserInfo(passportSerie, passportNumber, baseURL)

		if err != nil {
			log.Error("failed to request to info user", err)
			http.Error(w, "error request to other server", http.StatusBadRequest)
			return
		}

		id, err := userCreate.CreateUser(context, passportSerie, passportNumber, userInfoRes.Surname, userInfoRes.Name, userInfoRes.Patronymic, userInfoRes.Address)

		if err != nil {
			log.Error("failed to add user", err)
			http.Error(w, "not save user", http.StatusInternalServerError)
			return
		}

		log.Info("user added", slog.Int("id", id))

		responseOK(w, r, id)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Id: id,
	})
}
