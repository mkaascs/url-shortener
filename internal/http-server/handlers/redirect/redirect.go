package redirect

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/api/response"
	"url-shortener/internal/logging/sl"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

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
