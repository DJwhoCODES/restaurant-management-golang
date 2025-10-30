package routes

import (
	"github.com/djwhocodes/restaurant_management/controllers"
	"github.com/gin-gonic/gin"
)

func InvoiceRoutes(router *gin.Engine) {
	router.GET("/invoices", controllers.GetInvoices())
	router.GET("/invoices/:invoice_id", controllers.GetInvoice())
	router.POST("/invoices", controllers.CreateInvoice())
	router.PATCH("/invoices/:invoice_id", controllers.UpdateInvoice())
}
