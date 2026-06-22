package main

import (
	"api-gateway/handler"
	"api-gateway/middleware"
	"api-gateway/svc"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, svcCtx *svc.ServiceContext) {
	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestIDMiddleware())

	api := r.Group("/api/v1")
	{
		// ===== 库存接口（需认证）=====
		stock := api.Group("/stock")
		stock.Use(middleware.AuthMiddleware())
		{
			stock.POST("/batch-query", handler.BatchQueryStock(svcCtx))
		}

		// ===== 定时任务接口（需认证）=====
		cronjob := api.Group("/cronjob")
		cronjob.Use(middleware.AuthMiddleware())
		{
			cronjob.POST("/register", handler.RegisterTask(svcCtx))
			cronjob.POST("/unregister", handler.UnregisterTask(svcCtx))
			cronjob.POST("/list", handler.ListTasks(svcCtx))
			cronjob.POST("/set-enabled", handler.SetTaskEnabled(svcCtx))
			cronjob.POST("/executions", handler.ListExecutions(svcCtx))
			cronjob.POST("/trigger", handler.TriggerOnce(svcCtx))
			cronjob.POST("/retry", handler.RetryTask(svcCtx))
			cronjob.POST("/stats", handler.GetTaskStats(svcCtx))
		}

		// ===== 用户接口 =====
		user := api.Group("/user")
		{
			// 公开接口（无需认证）
			user.POST("/login", handler.Login(svcCtx))
			user.POST("/logout", handler.Logout(svcCtx))
			user.POST("/create", handler.CreateUser(svcCtx))
			user.PUT("/update", handler.UpdateUser(svcCtx))
			user.POST("/wx/login", handler.WxMiniProgramLogin(svcCtx))
			user.POST("/token/refresh", handler.RefreshToken(svcCtx))

			// 需认证接口
			auth := user.Group("")
			auth.Use(middleware.AuthMiddleware())
			{
				auth.POST("/list", handler.UserList(svcCtx))
				auth.POST("/info", handler.UserInfo(svcCtx))
			}
		}

		// ===== 订单接口 =====
		order := api.Group("/order")
		{
			order.POST("/create", handler.CreateOrder(svcCtx))
			order.POST("/commit-pay", handler.OrderCommitPay(svcCtx))
			order.GET("/state-check", handler.OrderStateCheck(svcCtx))
			order.POST("/cancel", handler.CancelOrder(svcCtx))
		}
	}
}
