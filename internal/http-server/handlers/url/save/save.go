package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/logging/sl"
	"url-shortener/internal/storage"
)

// TODO: move to config
const aliasLength = 6

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

func New(log *slog.Logger, saver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Error(err))
			responseError(w, r,
				http.StatusInternalServerError,
				"failed to decode request body")

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Error(err))
			responseValidationError(w, r, validateErr)
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := saver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			responseError(w, r,
				http.StatusConflict,
				"url already exists")

			return
		}

		if err != nil {
			log.Error("failed to save url", sl.Error(err))
			responseError(w, r,
				http.StatusInternalServerError,
				"failed to save url")

			return
		}

		log.Info("url saved", slog.String("url", req.URL), slog.Int64("id", id))
		responseCreated(w, r, alias)
	}
}

func responseCreated(w http.ResponseWriter, r *http.Request, alias string) {
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, Response{
		Response: response.Response{},
		Alias:    alias,
	})
}

func responseError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	render.Status(r, statusCode)
	render.JSON(w, r, Response{
		Response: response.Error(errorMessage),
	})
}

func responseValidationError(w http.ResponseWriter, r *http.Request, errors validator.ValidationErrors) {
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, response.ValidateError(errors))
}
