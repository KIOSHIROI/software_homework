package v1

import (
	"backend/middleware"
	"backend/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Home(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	// 今日必抢商品信息
	var commodity []models.Commodities
	models.DB.Where("deleted_at IS NULL").Order("sold DESC").Find(&commodity)
	data["commodityInfos"] = [][]models.Commodities{
		commodity[0:4], commodity[4:8],
	}
	// 分类商品信息
	var classification = map[string]string{"daily": "日常用品",
		"electronics": "电子产品", "books": "教材书籍", "clothes": "服装配饰", "sports": "体育用品",
		"healthy": "医疗健康", "entertainment": "娱乐体闲"}

	for k, v := range classification {
		var types = []string{}
		var temp []models.Commodities
		models.DB.Model(&models.Types{}).Where("firsts = ?", v).Select("seconds").Find(&types)
		models.DB.Where("types in (?) AND deleted_at IS NULL", types).Order("sold DESC").Find(&temp)
		data[k] = temp[0:5]
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func CommodityList(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	// 获取请求参数
	types := c.DefaultQuery("types", "")
	search := c.DefaultQuery("search", "")
	sort := c.DefaultQuery("sort", "")
	page := c.DefaultQuery("page", "1")
	p, _ := strconv.Atoi(page)
	// 商品分类列表
	var firsts = []string{}
	models.DB.Model(&models.Types{}).Distinct("firsts").Find(&firsts)
	var res []map[string]interface{}
	for _, f := range firsts {
		var seconds = []string{}
		models.DB.Model(&models.Types{}).Where("firsts = ?", f).Select("seconds").Find(&seconds)
		res = append(res, map[string]interface{}{"name": f, "value": seconds})
	}
	data["types"] = res
	// 商品列表信息
	var commodity []models.Commodities
	querys := models.DB.Model(&models.Commodities{}).Where("deleted_at IS NULL")
	if types != "" {
		querys = querys.Where("types = ?", types)
	}
	if sort != "" {
		querys = querys.Order(sort + "DESC")
	}
	if search != "" {
		querys = querys.Where("name like ?", "%"+search+"%")
	}
	querys, previous, next, count, pageCount := models.Paginate(querys, p)
	querys.Find(&commodity)
	data["commodityInfos"] = map[string]interface{}{
		"data": commodity, "previous": previous, "next": next,
		"count": count, "pageCount": pageCount,
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func CommodityDetail(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	id := c.Param("id")
	// 获取商品详情信息
	var commodity models.Commodities
	models.DB.Where("id = ?", id).First(&commodity)
	data["commodities"] = commodity
	// 获取推荐商品
	var recommend []models.Commodities
	models.DB.Where("id != ?", id).Order("sold DESC").Limit(5).Find(&recommend)
	data["recommend"] = recommend
	// 收藏状态
	data["likes"] = false
	// 获取请求头的Authorization
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		mc, _ := middleware.ParseToken(authHeader)
		if mc != nil {
			UserId := mc.UserId
			if UserId != 0 {
				var records []models.Records
				models.DB.Where("user_id = ? and commodity_id = ?", UserId, id).Find(&records)
				if len(records) > 0 {
					data["likes"] = true
				}
			}
		}
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

func CommodityCollect(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "请求失败"}
	data, _ := c.GetRawData()
	var body map[string]int64
	json.Unmarshal(data, &body)
	id := body["id"]
	userId, _ := c.Get("userId")
	var records []models.Records

	models.DB.Where("user_id = ? and commodity_id = ?", userId.(int64), id).Find(&records)

	if len(records) == 0 {
		models.DB.Model(&models.Commodities{}).Where("id = ?", id).Update("like", 1)
		r := models.Records{UserId: userId.(int64), CommodityId: id}
		models.DB.Create(&r)
		context["msg"] = "收藏成功"
		context["state"] = "success"
	} else {
		context["msg"] = "收藏取消"
		context["state"] = "success"
		models.DB.Unscoped().Delete(&records)
	}
	c.JSON(http.StatusOK, context)
}
