package sprava_krvi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (this *implDonorsAPI) GetDonors(ctx *gin.Context) {
	// ctx.AbortWithStatus(http.StatusNotImplemented)
	dummyData := gin.H{
		"message": "This is a dummy response",
	}
	ctx.JSON(http.StatusOK, dummyData)
}

func (this *implDonorsAPI) GetDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (this *implDonorsAPI) CreateDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

func (this *implDonorsAPI) UpdateDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
func (this *implDonorsAPI) DeleteDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
