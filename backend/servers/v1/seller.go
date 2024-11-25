package v1

import (
	"backend/models"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SellerAddCommodity(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "添加商品失败"}

	// // 从表单中获取文件
	// file, err := c.FormFile("img")
	// if err != nil {
	// 	print("v", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "获取图片文件失败", "details": err.Error()})
	// 	return
	// }

	// 指定文件保存路径，防止文件名冲突
	// savePath := "F:/commodity/images/" + time.Now().Format("20060102_150405_") + file.Filename

	// // 确保上传目录存在
	// dirPath := savePath[:4]
	// if _, err := os.Stat(dirPath); os.IsNotExist(err) {
	// 	err := os.MkdirAll(dirPath, 0755)
	// 	if err != nil {
	// 		print("v", err)
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建图片目录失败", "details": err.Error()})
	// 		return
	// 	}
	// }

	// // 保存文件到指定路径
	// if err := c.SaveUploadedFile(file, savePath); err != nil {
	// 	print("v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片文件失败", "details": err.Error()})
	// 	return
	// }

	// 解析其他表单数据
	data, err := c.GetRawData()
	if err != nil {
		print("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取请求数据失败", "details": err.Error()})
		return
	}

	var body map[string]interface{}
	err = json.Unmarshal(data, &body)
	if err != nil {
		print("%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "解析请求数据失败", "details": err.Error()})
		return
	}

	// 商品名称
	name, ok := body["name"].(string)
	if !ok {
		print("商品名称类型错误")
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品名称类型错误"})
		return
	}

	// 商品价格
	price, ok := body["price"].(float64)
	if !ok {
		print("商品价格类型错误")
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品价格类型错误"})
		return
	}

	// 商品详情
	details, ok := body["details"].(string)
	print("商品详情错误")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品详情类型错误"})
		return
	}

	// // 二级分类 ID
	// typeID, ok := body["type_id"].(float64)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "商品分类类型错误"})
	// 	return
	// }

	// 获取用户 ID
	sellerId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户ID失败"})
		return
	}

	// // 校验二级分类 ID 是否存在
	// var category models.Types
	// err = models.DB.First(&category, int(typeID)).Error
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类 ID"})
	// 	return
	// }

	// 创建商品对象
	commodity := models.Commodities{
		Name:     name,
		Price:    price,
		Sold:     0,
		Likes:    0,
		Details:  details,
		SellerId: sellerId.(int64),
		// Img:      savePath, // 图片路径
	}

	// 插入数据库
	err = models.DB.Create(&commodity).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加商品失败", "details": err.Error()})
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

	// 删除数据库中的记录
	err = models.DB.Delete(&commodity).Error
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
	context := gin.H{"msg": "获取一级分类失败", "state": "fail"}

	var firsts []string
	err := models.DB.Model(&models.Types{}).Distinct("firsts").Find(&firsts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询一级分类失败"})
		return
	}

	context["msg"] = "主分类获取成功"
	context["state"] = "success"
	data["firsts"] = firsts
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func GetTypesSeconds(c *gin.Context) {
	context := gin.H{"msg": "获取二级分类失败", "state": "fail"}
	data := gin.H{}
	first := c.Query("firsts")
	if first == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "一级分类不能为空"})
		return
	}
	var seconds []string
	err := models.DB.Model(&models.Types{}).Where("firsts = ?", first).Distinct("seconds").Find(&seconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询二级分类失败"})
		return
	}
	context["msg"] = "次分类获取成功"
	context["state"] = "success"
	data["seconds"] = seconds
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func GetMyCommodities(c *gin.Context) {
	context := gin.H{"msg": "获取商品失败", "state": "fail", "data": nil}

	// 获取卖家 ID
	sellerID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	// 查询数据库中该卖家的所有商品
	var commodities []models.Commodities
	err := models.DB.Where("seller_id = ?", sellerID).Find(&commodities).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询商品失败"})
		return
	}

	// 如果没有商品，返回空数组
	if len(commodities) == 0 {
		context["msg"] = "暂无商品"
		context["state"] = "success"
		context["data"] = []models.Commodities{}
		c.JSON(http.StatusOK, context)
		return
	}

	// 返回成功的响应
	context["msg"] = "商品获取成功"
	context["state"] = "success"
	context["data"] = commodities
	c.JSON(http.StatusOK, context)
}
