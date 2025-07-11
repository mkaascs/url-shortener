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

// Request for URL creation
type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

// Response with short URL and errors
type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

// New @Summary Create short URL
// @Description Converts long URL to short alias
// @Tags url
// @Accept json
// @Produce json
// @Param request body Request true "URL data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Failure 422 {object} Response
// @Router /url [post]
func New(log *slog.Logger, saver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"
		log = log.With(
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

		alias := request.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := saver.SaveURL(request.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", request.URL))
			response.RenderError(w, r,
				http.StatusConflict,
				"url already exists")
			return
		}

		if err != nil {
			log.Error("failed to save url", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to save url")
			return
		}

		log.Info("url saved", slog.String("url", request.URL), slog.Int64("id", id))
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
