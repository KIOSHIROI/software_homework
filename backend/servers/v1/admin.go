package v1

import (
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminDeleteUser 删除用户
func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.Users
	if err := models.DB.Where("id = ?", id).Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "用户删除成功",
	})
}

// AdminGetUsers 获取用户列表
func AdminGetUsers(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	var users []models.Users

	query := models.DB
	if search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "获取用户列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "获取用户列表成功",
		"data":  users,
	})
}

// AdminGetTypes 获取类型列表
func AdminGetTypes(c *gin.Context) {
	search := c.DefaultQuery("search", "")
	var types []models.Types

	query := models.DB
	if search != "" {
		query = query.Where("firsts LIKE ?", "%"+search+"%")
	}

	if err := query.Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "获取类型列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "获取类型列表成功",
		"data":  types,
	})
}

// AdminAddType 添加类型
func AdminAddType(c *gin.Context) {
	var newType models.Types
	if err := c.ShouldBindJSON(&newType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "fail",
			"msg":   "请求数据无效",
		})
		return
	}

	if err := models.DB.Create(&newType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "添加类型失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "类型添加成功",
		"data":  newType,
	})
}

// AdminModifyType 修改类型
func AdminModifyType(c *gin.Context) {
	id := c.Param("id")
	var updatedType models.Types

	if err := c.ShouldBindJSON(&updatedType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "fail",
			"msg":   "请求数据无效",
		})
		return
	}

	var existingType models.Types
	if err := models.DB.Where("id = ?", id).First(&existingType).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"state": "fail",
			"msg":   "类型不存在",
		})
		return
	}

	existingType.Firsts = updatedType.Firsts
	existingType.Seconds = updatedType.Seconds

	if err := models.DB.Save(&existingType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "修改类型失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "类型修改成功",
		"data":  existingType,
	})
}

// AdminDeleteType 删除类型
func AdminDeleteType(c *gin.Context) {
	id := c.Param("id")

	var t models.Types
	if err := models.DB.Where("id = ?", id).Delete(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"msg":   "删除类型失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "success",
		"msg":   "类型删除成功",
	})
}
