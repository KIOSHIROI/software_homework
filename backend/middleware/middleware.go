package middleware

import (
	"backend/models"
	"backend/settings"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	// 自行添加字段
	Username string `json:"username"`
	UserId   int64  `json:"userId"`
	// 内嵌JWT
	jwt.RegisteredClaims
}

// GenToken生成JWT
func GenToken(username string, userId int64) (string, error) {
	expire := time.Now().Add(settings.TokenExpireDuration)
	//Create our 'Claim'
	claims := CustomClaims{
		username,
		userId,
		jwt.RegisteredClaims{
			Issuer: "kioshiro",
		},
	}
	// 使用指定的签名方法创建签名对象
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(settings.Secret)
	//Token写入数据库
	j := models.Jwts{Token: token, Expire: expire}
	models.DB.Create(j)
	return token, err
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	// parse token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (i interface{}, err error) {
			return settings.Secret, nil
		})
	if err != nil {
		return nil, err
	}
	//check token
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
func JWTAuthMiddleware(c *gin.Context) {
	// Client Token types: 1. put in 'head'; 2. put in 'body'; 3. put in 'URL'
	// user type 1
	authHeader := c.Request.Header.Get("Authorozation")
	if authHeader == "" {
		c.JSON(http.StatusOK, gin.H{ // TODO: Why OK?
			"state": "fail",
			"msg":   "请求头的Authorization为空",
		})
		c.Abort()
		return
	}
	mc, err := ParseToken(authHeader) //myClaim
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"mag":   "无效的Token",
		})
		c.Abort()
		return
	}
	var jwts models.Jwts
	models.DB.Where("token = ?", authHeader).First(&jwts)
	if jwts.Token != "" {
		if jwts.Expire.After(time.Now()) {
			jwts.Expire = time.Now().Add(settings.TokenExpireDuration)
			models.DB.Save(&jwts)
		} else {
			// 删除表数据
			models.DB.Unscoped().Delete(&jwts)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg": "无效的Token",
		})
		c.Abort()
		return
	}
	// 将当前请求的username信息保存到请求的上下文c上
	// 路由函数通过处理c.Get("username")；哎获取当前请求的用户信息
	c.Set("userId", mc.UserId)
	c.Set("username", mc.Username)
	c.Next()
}
