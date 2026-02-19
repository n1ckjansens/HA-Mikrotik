package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/micro-ha/mikrotik-presence/addon/internal/http/handlers"
)

// NewRouter builds full HTTP routing tree for backend API and static frontend.
func NewRouter(api *handlers.API) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RecoverJSON)
	r.Use(middleware.Timeout(20 * time.Second))
	r.Use(StripIngressPrefix)
	r.Use(RequestLogger(api))

	r.Get("/healthz", api.Health)
	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/automation/action-types", api.ListActionTypes)
		apiRouter.Get("/automation/state-source-types", api.ListStateSourceTypes)

		apiRouter.Get("/automation/capabilities", api.ListCapabilities)
		apiRouter.Get("/automation/capabilities/{id}", func(w http.ResponseWriter, r *http.Request) {
			api.GetCapability(w, r, chi.URLParam(r, "id"))
		})
		apiRouter.Post("/automation/capabilities", api.CreateCapability)
		apiRouter.Put("/automation/capabilities/{id}", func(w http.ResponseWriter, r *http.Request) {
			api.UpdateCapability(w, r, chi.URLParam(r, "id"))
		})
		apiRouter.Delete("/automation/capabilities/{id}", func(w http.ResponseWriter, r *http.Request) {
			api.DeleteCapability(w, r, chi.URLParam(r, "id"))
		})
		apiRouter.Get("/automation/capabilities/{id}/devices", func(w http.ResponseWriter, r *http.Request) {
			api.ListCapabilityDevices(w, r, chi.URLParam(r, "id"))
		})
		apiRouter.Patch("/automation/capabilities/{id}/devices/{mac}", func(w http.ResponseWriter, r *http.Request) {
			api.PatchCapabilityDevice(w, r, chi.URLParam(r, "id"), chi.URLParam(r, "mac"))
		})
		apiRouter.Get("/global/capabilities", api.ListGlobalCapabilities)
		apiRouter.Patch("/global/capabilities/{capabilityId}", func(w http.ResponseWriter, r *http.Request) {
			api.PatchGlobalCapability(w, r, chi.URLParam(r, "capabilityId"))
		})

		apiRouter.Get("/devices", api.ListDevices)
		apiRouter.Get("/devices/{mac}/capabilities", func(w http.ResponseWriter, r *http.Request) {
			api.ListDeviceCapabilities(w, r, chi.URLParam(r, "mac"))
		})
		apiRouter.Patch("/devices/{mac}/capabilities/{capabilityId}", func(w http.ResponseWriter, r *http.Request) {
			api.PatchDeviceCapability(w, r, chi.URLParam(r, "mac"), chi.URLParam(r, "capabilityId"))
		})
		apiRouter.Get("/devices/{mac}", func(w http.ResponseWriter, r *http.Request) {
			api.GetDevice(w, r, chi.URLParam(r, "mac"))
		})
		apiRouter.Post("/devices/{mac}/register", func(w http.ResponseWriter, r *http.Request) {
			api.RegisterDevice(w, r, chi.URLParam(r, "mac"))
		})
		apiRouter.Patch("/devices/{mac}", func(w http.ResponseWriter, r *http.Request) {
			api.PatchDevice(w, r, chi.URLParam(r, "mac"))
		})
		apiRouter.Post("/refresh", api.Refresh)
	})

	r.Get("/*", api.Static)
	r.Get("/", api.Static)
	return r
}

// RunServer starts and gracefully stops HTTP server with context cancellation.
func RunServer(ctx context.Context, server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
