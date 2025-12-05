package router

import (
	"net/http"

	"github.com/dspo/go-homework/internal/application"
	"github.com/dspo/go-homework/internal/handler"
)

func NewRouter(app application.Context) http.Handler {
	mount(app)
	return app.Engine
}

func mount(app application.Context) {

	app.GET("/healthz", handler.HealthzHandler())

	app.GET("/api/audits")

	app.POST("/api/login")
	app.POST("/api/logout")
	app.GET("/api/me")
	app.PUT("/api/me")
	app.PUT("api/me/password")
	app.GET("/api/me/teams")
	app.DELETE("/api/me/teams/:team_id")
	app.GET("/api/me/projects")
	app.DELETE("/api/me/projects/:project_id")

	app.POST("/api/users")
	app.GET("/api/users")
	app.GET("/api/users/:user_id")
	app.DELETE("/api/users/:user_id")
	app.GET("/api/users/:user_id/teams")
	app.GET("/api/users/:user_id/projects")
	app.POST("/api/users/:user_id/roles")
	app.DELETE("/api/users/:user_id/roles/:role_id")

	app.GET("/api/teams")
	app.POST("/api/teams")
	app.GET("/api/teams/:team_id")
	app.PUT("/api/teams/:team_id")
	app.PATCH("/api/teams/:team_id")
	app.DELETE("/api/teams/:team_id")
	app.GET("/api/teams/:team_id/users")
	app.POST("/api/teams/:team_id/users")
	app.DELETE("/api/teams/:team_id/users/:user_id")
	app.GET("/api/teams/:team_id/projects")
	app.POST("/api/teams/:team_id/projects")

	app.GET("/api/projects/:project_id")
	app.PUT("/api/projects/:project_id")
	app.PATCH("/api/projects/:project_id")
	app.DELETE("/api/projects/:project_id")
	app.GET("/api/projects/:project_id/users")
	app.POST("/api/projects/:project_id/users")
	app.DELETE("/api/projects/:project_id/users/:user_id")

	app.GET("/api/roles")
	app.POST("/api/roles")
	app.DELETE("/api/roles/:role_id")

}
