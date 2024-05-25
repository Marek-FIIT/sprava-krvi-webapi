package sprava_krvi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Marek-FIIT/sprava-krvi-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateUnits - Creates new units
func (this *implUnitsAPI) CreateUnits(ctx *gin.Context) {
	sAmount := ctx.Query("amount")
	amount, err := strconv.Atoi(sAmount)
	if err != nil {
		message := "amount has to be integer"
		if sAmount == "" {
			message = "amount is required"
		}
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": message,
			},
		)
		return
	}

	var unit Unit
	if err := ctx.ShouldBindJSON(&unit); err != nil {
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

	if unit.Id == "" {
		// unit.Id = uuid.New().String()
		unit.DonationId = uuid.New().String()
		unit.Status = "unprocessed"
		unit.Frozen = false
		unit.Expiration = time.Now().AddDate(2, 0, 0)
		unit.CreatedAt = time.Now()
		unit.UpdatedAt = time.Now()
	}

	db, err := db_service.GetDbService[Unit](ctx, "db_service_units")
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

	// transaction, err := db.BeginTransaction(ctx)
	// if err != nil {
	// 	ctx.JSON(
	// 		http.StatusBadGateway,
	// 		gin.H{
	// 			"status":  "Bad Gateway",
	// 			"message": "failed to initialize transaction",
	// 			"error":   err.Error(),
	// 		})
	// 	return
	// }
	var ids []string
	var units []*Unit
	for i := 0; i < amount; i++ {
		unitCopy := unit
		unitCopy.Id = uuid.New().String()
		ids = append(ids, unitCopy.Id)
		units = append(units, &unitCopy)
		// err = transaction.CreateDocument(ctx, unitCopy.Id, &unitCopy)
		// switch err {
		// case nil:

		// case db_service.ErrConflict:
		// 	ctx.JSON(
		// 		http.StatusConflict,
		// 		gin.H{
		// 			"status":  "Conflict",
		// 			"message": "unit already exists",
		// 			"error":   err.Error(),
		// 		},
		// 	)
		// 	err = transaction.Rollback()
		// 	if err != nil {
		// 		log.Printf("transaction rollback failed")
		// 	}
		// 	return
		// default:
		// 	ctx.JSON(
		// 		http.StatusBadGateway,
		// 		gin.H{
		// 			"status":  "Bad Gateway",
		// 			"message": "Failed to create unit in database",
		// 			"error":   err.Error(),
		// 		},
		// 	)
		// 	err = transaction.Rollback()
		// 	if err != nil {
		// 		log.Printf("transaction rollback failed")
		// 	}
		// 	return
		// }
	}
	err = db.CreateDocuments(ctx, ids, units)
	switch err {
	case nil:

	case db_service.ErrConflict:
		ctx.JSON(
			http.StatusConflict,
			gin.H{
				"status":  "Conflict",
				"message": "A unit already exists",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create a unit in database",
				"error":   err.Error(),
			},
		)
		return
	}
	// err = transaction.Commit()
	// if err != nil {
	// 	ctx.JSON(
	// 		http.StatusBadGateway,
	// 		gin.H{
	// 			"status":  "Bad Gateway",
	// 			"message": "Failed to commit transaction",
	// 			"error":   err.Error(),
	// 		},
	// 	)
	// 	return
	// }
	ctx.JSON(
		http.StatusCreated,
		units,
	)
}

// DeleteUnit - Deletes the specific unit
func (this *implUnitsAPI) DeleteUnit(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// GetUnit - Provides the detail of the unit
func (this *implUnitsAPI) GetUnit(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// GetUnits - Provides the list of blood units
func (this *implUnitsAPI) GetUnits(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}

// UpdateUnit - updates the data of the specified unit
func (this *implUnitsAPI) UpdateUnit(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNotImplemented)
}
