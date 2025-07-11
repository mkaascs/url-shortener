package delete

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/logging/sl"
	"url-shortener/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	response.Response
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var request Request
		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Error("failed to decode request body", sl.Error(err))
			response.RenderError(w, r,
				http.StatusBadRequest,
				"failed to decode request body")
			return
		}

		log.Info("request body decoded", slog.Any("request", request))

		if err := validator.New().Struct(request); err != nil {
			var validateErrs validator.ValidationErrors
			errors.As(err, &validateErrs)

			log.Error("invalid request", sl.Error(err))
			response.RenderValidationError(w, r, validateErrs)
			return
		}

		err = deleter.DeleteURL(request.Alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url does not exist: %s", sl.Error(err))
			response.RenderError(w, r,
				http.StatusNotFound,
				"url with this alias does not exist")
			return
		}

		if err != nil {
			log.Error("failed to delete url", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to delete url")
			return
		}

		log.Info("url deleted", slog.String("alias", request.Alias))
		render.Status(r, http.StatusNoContent)
	}
}
