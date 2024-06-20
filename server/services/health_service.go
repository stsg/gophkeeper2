package services

import (
	"context"
	"log"
	"net/http"

	"github.com/stsg/gophkeeper2/server/repositories"
)

type HealthChecker struct {
	ctx context.Context
	db  repositories.DBProvider
}

// CheckDBHandler - check DB connection status
func (hc *HealthChecker) CheckDBHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := hc.db.HealthCheck(hc.ctx)
		if err != nil {
			log.Printf("failed db health check: %v", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}
}
