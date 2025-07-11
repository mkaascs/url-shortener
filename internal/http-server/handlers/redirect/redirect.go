package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/logging/sl"
	"url-shortener/internal/storage"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

type Response struct {
	response.Response
}

// New @Summary Redirect to URL
// @Description Redirect to URL by shorted alias
// @Tags redirect
// @Accept json
// @Produce json
// @Param alias path string true "Short URL alias"
// @Success 307
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /{alias} [get]
func New(log *slog.Logger, getter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.redirect.New"
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

		url, err := getter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url with alias not found")
			response.RenderError(w, r,
				http.StatusNotFound,
				"url with alias not found")
			return
		}

		if err != nil {
			log.Error("failed to get url", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to get url")
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
