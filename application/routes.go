package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CatalinPlesu/channel-service/handler"
	"github.com/CatalinPlesu/channel-service/repository/channel"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/channels", a.loadChannelRoutes)

	a.router = router
}

func (a *App) loadChannelRoutes(router chi.Router) {
	channelHandler := &handler.Channel{
		RdRepo: &channel.RedisRepo{
			Client: a.rdb,
		},
		PgRepo: channel.NewPostgresRepo(a.db),
	}

	router.Post("/", channelHandler.Create)
	router.Get("/", channelHandler.List)
	router.Get("/search/{name}", channelHandler.ListByName)
	router.Get("/{id}", channelHandler.GetByID)
	router.Put("/{id}", channelHandler.UpdateByID)
	router.Delete("/{id}", channelHandler.DeleteByID)
}
