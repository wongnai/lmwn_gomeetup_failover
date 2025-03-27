package health

import (
	"lmwn_gomeetup_failover/internal/db"
	"lmwn_gomeetup_failover/internal/memlimit"
	"lmwn_gomeetup_failover/internal/queue"
	"log"
	"net/http"
)

func RunHealthCheck(mongo *db.MongoDB, rabbitmq *queue.RabbitMQ) {
	srv := &http.Server{
		Addr: ":18080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isLowMem, err := memlimit.IsInLowMemory(memlimit.ProvideMemoryGetter(), true, 80)

			if (mongo != nil && !mongo.IsConnected()) ||
				(rabbitmq != nil && !rabbitmq.IsConnected()) ||
				(isLowMem || err != nil) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("{\"status\": \"unhealthy\"}"))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"status\": \"healthy\"}"))
		}),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()
}
