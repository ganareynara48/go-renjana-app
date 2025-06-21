package routes

import (
	"renjana-app/controllers"

	"renjana-app/middlewares"

	"github.com/gin-gonic/gin"
)

func MasterProductRoutes(r *gin.Engine) {
	mp := r.Group("/master_product", middlewares.AuthMiddleware())
	{
		mp.POST("/insert", controllers.InsertMasterProduct)
		mp.GET("/all", controllers.GetAllProduct)
		mp.GET("/first", controllers.GetProductByID)
		mp.POST("/update", controllers.UpdateProduct)
		mp.POST("/delete", controllers.DeleteProduct)
	}
}
