package object

import "github.com/gin-gonic/gin"

func RegisterObjectRoutes(router *gin.Engine) {
	objectGroup := router.Group("/object")
	{
		objectGroup.POST("/stage", HandleStage)
	}
}
