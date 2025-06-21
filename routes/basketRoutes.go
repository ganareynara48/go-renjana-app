package routes

import (
	"renjana-app/controllers"
	"renjana-app/middlewares"

	"github.com/gin-gonic/gin"
)

func BasketRoutes(r *gin.Engine) {
	basket := r.Group("/basket", middlewares.AuthMiddleware())
	{
		basket.POST("/insert", controllers.InsertBasket)
		basket.POST("/add_qty", controllers.AddQty)
		basket.POST("/remove_qty", controllers.RemoveQty)
	}
}
