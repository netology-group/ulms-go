package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/handlers"
	auth "github.com/netology-group/ulms-auth-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func (app *App) setRoutes() {
	app.router.PathPrefix(`/swagger-ui/`).Handler(http.StripPrefix("/swagger-ui", http.FileServer(http.Dir("/swagger-ui"))))

	app.router.Handle(`/metrics`, promhttp.Handler())

	api := app.router.PathPrefix(`/api/v1/{audience:[\w\-\.]+}`).Subrouter()
	api.Use(handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(&recoveryLogger{}),
	))
	api.Use(bodyCloseMiddleware)
	api.Use(handlers.CORS(app.corsOptions()...))
	api.Use(app.auth.TokenValidationMiddleware())
}

func (app *App) corsOptions() []handlers.CORSOption {
	return []handlers.CORSOption{
		handlers.AllowCredentials(),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Random-Id"}),
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch}),
		handlers.AllowedOrigins(app.Config.CORS.AllowedOrigins),
		handlers.MaxAge(app.Config.CORS.MaxAge),
	}
}

func bodyCloseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}
		next.ServeHTTP(w, r)
	})
}

type recoveryLogger struct{}

func (logger *recoveryLogger) Println(args ...interface{}) {
	logrus.Error(args...)
}

// DecodeForm decodes data from request body to the specified form
func DecodeForm(form validation.Validatable, r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType+";", "application/json;") {
		if err := json.NewDecoder(r.Body).Decode(form); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown content type: %v", contentType)
	}
	return form.Validate()
}

// RenderResponse renders response in JSON format using provided structure (object)
// If err != nil then panic will be thrown instead
// Note: several well-known errors will be rendered with appropriate HTTP status codes without panic:
//   sql.ErrNoRows           - 404
//   auth.ErrorNotAuthorized - 401
func RenderResponse(w http.ResponseWriter, httpStatusCode int, object interface{}, err error) {
	switch err {
	case nil:
		if object != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(httpStatusCode)
			body, err := json.Marshal(object)
			if err != nil {
				panic(err.Error())
			}
			_, _ = w.Write(body)
		} else {
			w.WriteHeader(httpStatusCode)
		}
	case sql.ErrNoRows:
		http.Error(w, "not found", http.StatusNotFound)
	case auth.ErrorNotAuthorized:
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		panic(err.Error())
	}
}
