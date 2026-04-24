package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
为了方便前后端联合调试，这里所有的网络状态码返回结果都是http.StatusOK即200。通过gin框架中gin.H结构体返回具体的code和message。
前端按照相应的code将message在f12控制台打印或者回显到页面中。ret = 0时是业务走完正常流程返回的结果码，状态码统一为200；
ret = -1是系统错误，比如序列化失败，redis缓存失败等，状态码统一为500；ret = -2是业务数据问题导致未正常走完业务流程返回的结果码，状态码统一为400
*/
func JsonBack(ctx *gin.Context, msg string, ret int, data any) {
	switch ret {
	case 0: // 业务正常走完流程返回的结果码
		if data != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": msg,
				"data":    data,
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": msg,
			})
		}
	case -1: // 系统错误导致未正常走完业务流程返回的结果码
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusInternalServerError,
			"message": msg,
		})
	case -2: // 业务数据问题导致未正常走完业务流程返回的结果码
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusBadRequest,
			"message": msg,
		})
	}

}
