package httpapi

import (
	"log"
	"net/http"
	"time"
)

// 简单的访问日志中间件
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("completed in %v", time.Since(start))
	})
}
