package handler

import (
	"net/http"

	"github.com/jabuxas/abyss/internal/middleware"
)

// SetupRoutes configures the HTTP routes for the application.
func (h *Handler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", middleware.LogHandler(h.App.Logger, h.IndexHandler))

	mux.HandleFunc("/raw/", middleware.LogHandler(h.App.Logger, h.ServeRawFileHandler))

	mux.Handle("/tree/", http.StripPrefix("/tree",
		middleware.LogHandler(h.App.Logger,
			middleware.BasicAuth(h.App, h.ListAllFilesHandler),
		),
	))
	mux.HandleFunc("/tree", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tree/", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/last", middleware.LogHandler(h.App.Logger,
		middleware.BasicAuth(h.App, h.LastUploadedHandler),
	))

	mux.HandleFunc("/token", middleware.LogHandler(h.App.Logger,
		middleware.BasicAuth(h.App, h.CreateJWTHandler),
	))
}
