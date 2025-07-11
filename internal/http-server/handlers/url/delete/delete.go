package delete

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/logging/sl"
	"url-shortener/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type Response struct {
	response.Response
}

// New @Summary Delete URL
// @Description Delete URL with shorted alias
// @Tags url
// @Accept json
// @Produce json
// @Param alias path string true "Short URL alias"
// @Success 204
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /url/{alias} [delete]
func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"
		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("no alias provided")
			response.RenderError(w, r,
				http.StatusBadRequest,
				"no alias provided")
			return
		}

		err := deleter.DeleteURL(alias)
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

		log.Info("url deleted", slog.String("alias", alias))
		render.Status(r, http.StatusNoContent)
	}
}
