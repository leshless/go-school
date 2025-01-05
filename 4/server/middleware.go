package server

import (
	"net/http"
)

const (
	AccessToken = "top_secret"
	AccessTokenKey = "AccessToken"
)

func AuthorizeMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		accessToken := req.Header.Get(AccessTokenKey)

		if accessToken != AccessToken {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(res, req)
	})
}