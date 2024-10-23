package routers

import (
	"backend/middleware"
	v1 "backend/servers/v1"
	"backend/settings"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(settings.Mode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.StaticFS("/static", http.Dir("static"))
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // 允许所有来源
		// 或者只允许特定来源
		// AllowOrigins: []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	// 定义路由
	apiv1 := r.Group("api/v1/")

	commodity := apiv1.Group("")
	{
		// 网站首页
		commodity.GET("home/", v1.Home)
		// 商品列表
		commodity.GET("commodity/list/", v1.CommodityList)
		// 商品详细
		commodity.GET("commodity/detail/:id/", v1.CommodityDetail)
		// 用户注册登录
		commodity.POST("shopper/login/", v1.ShopperLogin)
	}
	shopper := apiv1.Group("")
	shopper.Use(middleware.JWTAuthMiddleware)
	{
		// 商品收藏
		shopper.POST("commodity/collect/", v1.CommodityCollect)
		// 退出登录
		shopper.POST("commodity/logout/", v1.ShopperLogout)
		// 个人主页
		shopper.GET("shopper/home/", v1.ShopperHome)
		// 加入购物车
		shopper.POST("shopper/shopcart", v1.ShopperShopCart)
		// 购物车列表
		shopper.GET("shopper/shopcart", v1.ShopperShopCart)
		// 在线支付
		shopper.POST("shopper/pays/", v1.ShopperPays)
		// 删除购物车商品
		shopper.POST("shopper/delete/", v1.ShopperDelete)
	}

	return r
}
