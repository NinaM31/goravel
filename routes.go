package goravel

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (grvl *Goravel) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)

	if grvl.Debug {
		mux.Use(middleware.Logger)
	}
	mux.Use(middleware.Recoverer)

	return mux
}
