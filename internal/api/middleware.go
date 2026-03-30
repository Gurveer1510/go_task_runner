package api

import (
	"net/http"

	"github.com/go-task-runner/internal/logger"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		defer func(){
			if err := recover(); err != nil {
				logger.Log.Error("panic recovered", "error", err)
				http.Error(rw, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(rw, r)
	})
}