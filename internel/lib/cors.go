package lib

import (
	"net/http"

	"github.com/rs/cors"
)

func Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodOptions, http.MethodPut, http.MethodPost, http.MethodGet},
		AllowedHeaders: []string{
			"Content-Length",
			"Referer",
			"User-Agent",
			"X-Api-Client",
			"X-Api-Trace",
			"X-Api-Version",
			"X-Api-Token",
			"Content-Type",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Api-Client",
			"X-Api-Trace",
			"X-Api-Version",
			"Content-Type",
		},
		MaxAge:             43200, //time.hour * 12
		AllowCredentials:   false,
		OptionsPassthrough: true,
		Debug:              false,
	})
}
