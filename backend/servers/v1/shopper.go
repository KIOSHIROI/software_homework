package v1

import (
	"backend/middleware"
	"backend/models"
	"backend/settings"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"golang.org/x/crypto/bcrypt"
)

func ShopperLogin(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "注册或登录失败"}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&body); err != nil {
		context["msg"] = "请求数据无效"
		c.JSON(http.StatusBadRequest, context)
		return
	}

	username := body.Username
	password := body.Password
	if username != "" && password != "" {
		// 生成登录时间
		lastLogin := time.Now()
		context["last_login"] = lastLogin.Format("2006-01-02 15:04:05")

		// 查找用户
		var user models.Users
		result := models.DB.Where("username = ?", username).First(&user)

		if result.Error == nil {
			// 用户存在，验证密码
			fmt.Println("username exists.")
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

			if err != nil {
				context["msg"] = "请输入正确密码"
			} else {
				fmt.Println("successful login.")
				// 登录成功，更新最后登录时间
				user.LastLogin = lastLogin
				models.DB.Save(&user)

				context["state"] = "success"
				context["msg"] = "登录成功"
				context["result"] = true
			}
		} else {
			newUser := models.Users{
				Username:  username,
				Password:  password,
				IsStaff:   1,
				LastLogin: lastLogin,
			}

			if err := models.DB.Create(&newUser).Error; err != nil {
				context["msg"] = "注册失败"
				c.JSON(http.StatusInternalServerError, context)
				return
			}

			context["state"] = "success"
			context["msg"] = "注册成功"
		}

		// 创建Token
		if user.ID > 0 {
			token, err := middleware.GenToken(username, int64(user.ID))
			if err == nil {
				context["token"] = token
			}
		}
	}

	c.JSON(http.StatusOK, context)
}

func ShopperShopCart(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	userId, _ := c.Get("userId")
	if c.Request.Method == "GET" {
		if userId != 0 {
			var carts []models.Carts
			models.DB.Preload("Commodities").Where("user_id = ?",
				userId.(int64)).Order("id DESC").Find(&carts)
			context["data"] = carts
		}
	}
	if c.Request.Method == "POST" {
		context = gin.H{"state": "fail", "msg": "获取失败"}
		var body struct {
			Id       int64 `json:"id"`
			Quantity int64 `json:"quantity"`
		}
		c.BindJSON(&body)
		id := body.Id
		quantity := body.Quantity
		var commodity models.Commodities
		models.DB.Where("id = ?", id).First(&commodity)
		if commodity.ID > 0 {
			var cart models.Carts
			models.DB.Where("commodity_id = ? and user_id = ?", id, userId).Find(&cart)
			if cart.ID > 0 {
				cart.Quantity += quantity
				models.DB.Save(&cart)
			} else {
				cart := models.Carts{UserId: userId.(int64), CommodityId: id, Quantity: quantity}
				models.DB.Create(&cart)
			}
			context = gin.H{"state": "success", "msg": "加购成功"}
		}
	}
	c.JSON(http.StatusOK, context)
}

// 购物车功能
func ShopperDelete(c *gin.Context) {
	var body struct {
		CartId int64 `json:"cartId"`
	}
	c.BindJSON(&body)
	cartId := body.CartId
	var cart []models.Carts
	if cartId != 0 {
		models.DB.Where("id = ?", cartId).Find(&cart)
	} else {
		userId, _ := c.Get("userId")
		models.DB.Where("user_id = ?", userId).Find(&cart)
	}
	models.DB.Unscoped().Delete(&cart)
	context := gin.H{"state": "success", "msg": "删除成功"}
	c.JSON(http.StatusOK, context)
}

func ShopperPays(c *gin.Context) {
	var body struct {
		Total   string   `json:"total"`
		PayInfo string   `json:"payInfo"`
		CartID  []string `json:"cartId"`
	}
	c.BindJSON(&body)
	total := strings.Replace(body.Total, "￥", "", -1)
	payInfo := body.PayInfo
	cartId := body.CartID
	if total == "" {
		context := gin.H{"state": "fail", "msg": "支付失败，请输入金额"}
		c.JSON(http.StatusOK, context)
	}
	if payInfo == "" {
		payInfo = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	userId, _ := c.Get("userId")
	var order models.Orders
	if order.ID == 0 {
		carts := models.Orders{UserId: userId.(int64), Price: total, PayInfo: payInfo, State: 0}
		models.DB.Create(&carts)
	}
	if len(cartId) != 0 {
		models.DB.Unscoped().Delete(&[]models.Carts{}, cartId)
	}
	client, _ := alipay.New(settings.AppId, settings.AppPrivateKeyString, false)
	client.LoadAliPayPublicKey(settings.AlipayPublicKeyString)
	var p = alipay.TradePagePay{}
	p.ReturnURL = "http://localhost:8010/#/shopper"
	p.Body = "支付宝测试"
	p.Subject = "测试test"
	p.OutTradeNo = payInfo
	p.TotalAmount = total
	p.ProductCode = "FAST_INSTANT_TRACE_PAY"
	url, _ := client.TradePagePay(p)
	payURL := url.String()
	context := gin.H{"state": "success", "msg": "支付成功", "data": payURL}
	fmt.Println(payURL)
	c.JSON(http.StatusOK, context)
}

func ShopperLogout(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		// 没有Token，直接视为退出成功
		c.JSON(http.StatusOK, gin.H{"state": "success", "msg": "退出成功"})
		return
	}

	var jwts models.Jwts
	result := models.DB.Where("token = ?", authHeader).First(&jwts)
	if result.Error != nil {
		// 查询过程中出现错误
		c.JSON(http.StatusInternalServerError, gin.H{"state": "fail", "msg": "查找Token失败"})
		return
	}

	if result.RowsAffected == 0 {
		// Token未找到，但仍视为退出成功
		c.JSON(http.StatusOK, gin.H{"state": "success", "msg": "退出成功"})
		return
	}

	// 删除找到的Token
	if err := models.DB.Unscoped().Delete(&jwts).Error; err != nil {
		// 删除Token失败
		c.JSON(http.StatusInternalServerError, gin.H{"state": "fail", "msg": "删除Token失败"})
		return
	}

	// Token删除成功，返回成功响应
	c.JSON(http.StatusOK, gin.H{"state": "success", "msg": "退出成功"})
}

func ShopperHome(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	userId, _ := c.Get("userId")
	payInfo := c.DefaultQuery("out_trade_no", "")
	if payInfo != "" {
		models.DB.Model(&models.Orders{}).Where("pay_info = ?", payInfo).Update("state", 1)
	}
	if userId != 0 {
		var orders []models.Orders
		models.DB.Where("user_id = ? AND deleted_at IS NULL", userId).Order("id DESC").Find(&orders)
		data["orders"] = orders
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func DeleteOrder(c *gin.Context) {
	// 默认的响应内容
	context := gin.H{"state": "fail", "msg": "请求失败"}

	// 获取 orderId，进行类型断言
	orderId, exists := c.Get("orderId")
	if !exists || orderId == nil {
		c.JSON(http.StatusBadRequest, gin.H{"state": "fail", "msg": "orderId 不存在"})
		return
	}

	// 确保 orderId 是整数类型，进行类型断言
	orderIdInt, ok := orderId.(int)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"state": "fail", "msg": "orderId 类型错误"})
		return
	}

	// 执行软删除
	result := models.DB.Model(&models.Orders{}).Where("id = ?", orderIdInt).Update("deleted_at", time.Now())
	if result.Error != nil {
		context = gin.H{"state": "fail", "msg": "删除失败", "error": result.Error.Error()}
	} else if result.RowsAffected > 0 {
		context = gin.H{"state": "success", "msg": "删除成功"}
	} else {
		context = gin.H{"state": "fail", "msg": "未找到该订单"}
	}
	c.JSON(http.StatusOK, context)
}
