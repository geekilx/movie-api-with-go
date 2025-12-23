package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) route() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)

	// endpoint that require authentication in order to get visited
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))

	// no authentication needed
	router.HandlerFunc(http.MethodPost, "/v1/users", app.createUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users/authentication", app.userAuthenticationHandler)

	return app.panicRecover(app.rateLimit(app.Authenticate(router)))

}
