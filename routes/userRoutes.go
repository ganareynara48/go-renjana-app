package routes

import (
	"renjana-app/controllers"

	"renjana-app/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	users := r.Group("/users", middlewares.AuthMiddleware())
	{
		// Get login user info
		users.GET("/info", controllers.InfoUser)

		// Get all users
		users.GET("/all", controllers.GetAllUser)

		// Get user by ID
		users.GET("/first", controllers.GetUserByID)

		// Update user by ID
		users.POST("/update", controllers.UpdateUser)

		// Delete user
		users.POST("/delete", controllers.DeleteUser)

		// Group: /users/address
		address := users.Group("/address")
		{
			address.POST("/insert", controllers.InsertAddress)
			address.POST("/update", controllers.UpdateAddress)
			address.GET("/get_all", controllers.GetAllAddressByUserID)
			address.GET("/get_active", controllers.GetActiveAddress)
			address.POST("/delete", controllers.DeleteAddress)
		}
	}
}
