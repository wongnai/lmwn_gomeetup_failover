package http

import (
	"context"
	"lmwn_gomeetup_failover/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(svc *service.Service) *HTTPServer {
	r := gin.Default()
	r.Use(panicRecoveryMiddleware()) // Apply panic recovery middleware

	r.POST("/create-order", CreateOrderHandler(svc))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	return &HTTPServer{server: srv}
}

func panicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			}
		}()
		c.Next()
	}
}

func (h *HTTPServer) Start() {
	log.Println("Starting HTTP server on port 8080")
	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (h *HTTPServer) Stop(ctx context.Context) {
	log.Println("Shutting down HTTP server...")
	if err := h.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}
	log.Println("HTTP server shutdown complete.")
}
