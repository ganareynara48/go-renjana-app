package routes

import (
	"renjana-app/controllers"

	"renjana-app/middlewares"

	"github.com/gin-gonic/gin"
)

func ProductCategoryRoutes(r *gin.Engine) {
	pc := r.Group("/product_category", middlewares.AuthMiddleware())
	{
		//input produc category
		pc.POST("/insert", controllers.InsertCategory)
		pc.GET("/get_all", controllers.GetAllCategory)
		pc.GET("/get_by_id", controllers.GetCategoryByID)
		pc.POST("/update", controllers.UpdateCategory)
		pc.POST("/delete", controllers.DeleteCategory)
	}
}
