package sprava_krvi

import (
	"net/http"
	"time"

	"github.com/Marek-FIIT/sprava-krvi-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	// ctx.AbortWithStatus(http.StatusNotImplemented)
	var donor Donor

	if err := ctx.ShouldBindJSON(&donor); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			},
		)
		return
	}

	db, err := db_service.GetDbService[Donor](ctx)
	if err != nil {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "failed to access db_service",
				"error":   err.Error(),
			})
		return
	}

	if donor.Id == "" {
		donor.Id = uuid.New().String()
		donor.CreatedAt = time.Now()
		donor.UpdatedAt = time.Now()
	}

	err = db.CreateDocument(ctx, donor.Id, &donor)
	switch err {
	case nil:
		ctx.JSON(
			http.StatusCreated,
			donor,
		)
	case db_service.ErrConflict:
		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "donor already exists",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create donor in database",
				"error":   err.Error(),
			},
		)
	}
}

func (this *implDonorsAPI) UpdateDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
func (this *implDonorsAPI) DeleteDonor(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
