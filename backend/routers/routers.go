package routers

import (
	"backend/middleware"
	v1 "backend/servers/v1"
	"backend/servers/v1/chats"
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
	r.StaticFS("/images", http.Dir("F:/mall/commodity/images/"))
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))

	// 定义路由
	apiv1 := r.Group("api/v1/")

	chat := chats.NewChat()
	go chat.Run()

	r.LoadHTMLFiles("index.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/chat/", func(ctx *gin.Context) {
		chats.CreateWs(ctx, chat)
	})

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
		commodity.POST("shopper/logout/", v1.ShopperLogout)
	}
	shopper := apiv1.Group("")
	shopper.Use(middleware.JWTAuthMiddleware)
	{
		// 商品收藏
		shopper.POST("commodity/collect/", v1.CommodityCollect)
		// 退出登录
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
		shopper.POST("shopper/order/delete", v1.DeleteOrder)
	}

	seller := apiv1.Group("seller/")
	seller.Use(middleware.JWTAuthMiddleware)
	{
		// 添加商品
		seller.POST("add/commodity", v1.SellerAddCommodity)
		seller.POST("delete/commodity", v1.SellerDeleteCommodity)
		seller.POST("modify/commodity", v1.SellerModifyCommodity)
		seller.GET("types/firsts", v1.GetTypesFirsts)
		seller.GET("types/seconds", v1.GetTypesSeconds)

	}

	admin := apiv1.Group("admin/")
	{
		admin.GET("types", v1.AdminGetTypes)
		admin.POST("add/type", v1.AdminAddType)
		admin.POST("delete/type", v1.AdminDeleteType)
		admin.POST("modify/type", v1.AdminModifyType)

		admin.GET("users", v1.AdminGetUsers)
		admin.POST("delete/user")
	}
	return r
}
