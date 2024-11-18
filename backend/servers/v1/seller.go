package v1

import (
	"backend/models"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetMyCommodities(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "获取商品失败"}
	data := gin.H{}
	sellerId, _ := c.Get("userId")

	// 确保sellerId是有效的
	if sellerId == nil {
		context["msg"] = "无效的卖家ID"
		c.JSON(http.StatusBadRequest, context)
		return
	}

	var commodities []models.Commodities
	// 根据sellerId查询商品，同时排除软删除的商品
	err := models.DB.Where("seller_id = ? AND deleted_at IS NULL", sellerId).Find(&commodities).Error
	if err != nil {
		context["msg"] = "查询商品失败"
		c.JSON(http.StatusInternalServerError, context)
		return
	}

	data["commodities"] = commodities
	context["state"] = "success"
	context["msg"] = "获取商品成功"
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func SellerAddCommodity(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "添加商品失败"}

	// 从表单中获取文件
	file, err := c.FormFile("img")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取图片文件失败"})
		return
	}

	// 指定文件保存路径
	savePath := "F:/commodity/images/" + file.Filename

	// 确保上传目录存在
	if _, err := os.Stat(savePath[:4]); os.IsNotExist(err) {
		err := os.MkdirAll(savePath[:4], 0755)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建图片目录失败"})
			return
		}
	}

	// 保存文件到指定路径
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片文件失败"})
		return
	}

	// 解析其他表单数据
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取请求数据失败"})
		return
	}

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求数据失败"})
		return
	}

	name, ok := body["name"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品名称类型错误"})
		return
	}

	// commodityType, ok := body["type"].(string)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "商品类型类型错误"})
	// 	return
	// }

	price, ok := body["price"].(float64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品价格类型错误"})
		return
	}

	// discount, ok := body["discount"].(float64)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "商品折扣类型错误"})
	// 	return
	// }

	// stock, ok := body["stock"].(int64)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "商品库存类型错误"})
	// 	return
	// }

	// img, ok := body["img"].(string)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传错误"})
	// 	return
	// }

	details, ok := body["details"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品详情错误"})

		return
	}
	sellerId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户ID失败"})
		return
	}

	// 假设 models.Commodities 已经定义，并且 img 和 details 也已经赋值
	commodity := models.Commodities{
		Name: name,
		// Types:    commodityType,
		Price: price,
		Sold:  0,
		Likes: 0,
		// Discount: discount,
		// Stock:    stock,
		// Img:      img,
		Created:  time.Now(),
		Details:  details,
		SellerId: sellerId.(int64),
	}

	err = models.DB.Preload("Seller").Create(&commodity).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加商品失败"})
		return
	}

	context["state"] = "success"
	context["msg"] = "商品添加成功"
	c.JSON(http.StatusOK, context)
}

func SellerDeleteCommodity(c *gin.Context) {
	// 获取商品ID
	commodityIDStr := c.Param("id")
	commodityID, err := strconv.Atoi(commodityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
		return
	}

	// 在数据库中查找商品
	var commodity models.Commodities
	err = models.DB.Preload("Seller").First(&commodity, commodityID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查找商品失败"})
		return
	}

	// 执行软删除
	err = models.DB.Model(&commodity).Update("DeletedAt", time.Now()).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除商品失败"})
		return
	}

	// 删除文件系统中的图片文件
	imgPath := "F:/commodity/images/" + commodity.Img
	if _, err := os.Stat(imgPath); err == nil {
		err = os.Remove(imgPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除图片文件失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"msg": "商品删除成功"})
}

func SellerModifyCommodity(C *gin.Context) {

}

func GetTypesFirsts(c *gin.Context) {
	data := gin.H{}
	context := gin.H{"msg": "获取失败", "state": "fail"}
	var firsts []string
	err := models.DB.Model(&models.Types{}).Distinct("firsts").Find(&firsts)
	if err != nil {
		context["msg"] = "主分类获取成功"
		context["state"] = "success"
	}
	data["firsts"] = firsts
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func GetTypesSeconds(c *gin.Context) {
	context := gin.H{"msg": "获取失败", "state": "fail"}
	data := gin.H{}
	first := c.Query("first")
	var seconds []string
	err := models.DB.Model(&models.Types{}).Where("firsts = ?", first).Distinct("seconds").Find(&seconds)
	if err != nil {
		context["msg"] = "次分类获取成功"
		context["state"] = "success"
	}
	data["seconds"] = seconds
	context["data"] = data
	c.JSON(http.StatusOK, context)
}
