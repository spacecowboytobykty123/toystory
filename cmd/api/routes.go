package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.requireActivatedUser(app.healthCheckHandler))
	router.HandlerFunc(http.MethodPost, "/v1/toy", app.requirePermission("toys:write", app.createToyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/toy/:id", app.requirePermission("toys:read", app.showToyHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/toy/:id", app.requirePermission("toys:write", app.updateToyHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/toy/:id", app.requirePermission("toys:write", app.deleteToyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/toys", app.requirePermission("toys:read", app.listToysHandler))

	router.HandlerFunc(http.MethodPost, "/v1/toy/:id/comment", app.requirePermission("toys:comment", app.createCommentHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))

}
