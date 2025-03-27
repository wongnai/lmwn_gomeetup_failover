package http

import (
	"net/http"

	"lmwn_gomeetup_failover/internal/service"

	"github.com/gin-gonic/gin"
)

func CreateOrderHandler(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Query("orderID")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}
		svc.CreateOrder(orderID)
		c.JSON(http.StatusOK, gin.H{"message": "Order created", "orderID": orderID})
	}
}
