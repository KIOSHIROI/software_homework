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

// GenToken 生成 JWT
func GenToken(username string, userId int64) (string, error) {
	expire := time.Now().Add(settings.TokenExpireDuration)
	claims := CustomClaims{
		Username: username,
		UserId:   userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "kioshiro",
			ExpiresAt: jwt.NewNumericDate(expire), // 设置过期时间
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(settings.Secret)
	if err != nil {
		return "", err // 返回错误
	}

	// Token写入数据库
	j := models.Jwts{Token: token, Expire: expire}
	if err := models.DB.Create(&j).Error; err != nil {
		return "", err // 返回数据库错误
	}
	return token, nil
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
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"state": "fail",
			"msg":   "请求头的Authorization为空",
		})
		c.Abort()
		return
	}

	mc, err := ParseToken(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"state": "fail",
			"msg":   "无效的Token",
		})
		c.Abort()
		return
	}

	var jwts models.Jwts
	result := models.DB.Where("token = ?", authHeader).First(&jwts)
	if result.Error != nil || jwts.Token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"state": "fail",
			"msg":   "无效的Token",
		})
		c.Abort()
		return
	}

	if jwts.Expire.After(time.Now()) {
		// 如果 token 有效，刷新过期时间
		jwts.Expire = time.Now().Add(settings.TokenExpireDuration)
		if err := models.DB.Save(&jwts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"state": "fail",
				"msg":   "更新Token失败",
			})
			c.Abort()
			return
		}
	} else {
		// Token 已过期，删除表数据
		models.DB.Unscoped().Delete(&jwts)
		c.JSON(http.StatusUnauthorized, gin.H{
			"state": "fail",
			"msg":   "Token已过期",
		})
		c.Abort()
		return
	}

	c.Set("userId", mc.UserId)
	c.Set("username", mc.Username)
	c.Next()
}
