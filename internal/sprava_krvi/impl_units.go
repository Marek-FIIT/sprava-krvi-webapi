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

// GetUnit - Provides the detail of the unit
func (this *implUnitsAPI) GetUnit(ctx *gin.Context) {
	unitId := ctx.Param("unitId")
	if unitId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Unit ID is required",
			},
		)
		return
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

	unit, err := db.FindDocument(ctx, unitId)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusOK,
			unit,
		)
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Unit not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load unit from database",
				"error":   err.Error(),
			})
		return
	}
}

// GetUnits - Provides the list of blood units
func (this *implUnitsAPI) GetUnits(ctx *gin.Context) {
	filters := make(map[string]interface{})
	var filterErrs []error
	if bloodType := ctx.Query("bloodType"); bloodType != "" {
		filters["bloodtype"] = bloodType
	}
	if bloodRh := ctx.Query("bloodRh"); bloodRh != "" {
		filters["bloodrh"] = bloodRh
	}
	if status := ctx.Query("status"); status != "" {
		filters["status"] = status
	}
	if location := ctx.Query("location"); location != "" {
		filters["location"] = location
	}
	if erythrocytes := ctx.Query("erythrocytes"); erythrocytes != "" {
		erythrocytesBool, err := strconv.ParseBool(erythrocytes)
		filterErrs = append(filterErrs, err)
		filters["contents.erythrocytes"] = erythrocytesBool
	}
	if leukocytes := ctx.Query("leukocytes"); leukocytes != "" {
		leukocytesBool, err := strconv.ParseBool(leukocytes)
		filterErrs = append(filterErrs, err)
		filters["contents.leukocytes"] = leukocytesBool
	}
	if platelets := ctx.Query("platelets"); platelets != "" {
		plateletsBool, err := strconv.ParseBool(platelets)
		filterErrs = append(filterErrs, err)
		filters["contents.platelets"] = plateletsBool
	}
	if plasma := ctx.Query("plasma"); plasma != "" {
		plasmaBool, err := strconv.ParseBool(plasma)
		filterErrs = append(filterErrs, err)
		filters["contents.plasma"] = plasmaBool
	}
	if frozen := ctx.Query("frozen"); frozen != "" {
		frozenBool, err := strconv.ParseBool(frozen)
		filterErrs = append(filterErrs, err)
		filters["frozen"] = frozenBool
	}
	for _, err := range filterErrs {
		if err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{
					"status":  http.StatusBadRequest,
					"message": "Could not parse filters",
				},
			)
			return
		}
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

	units, err := db.FindDocuments(ctx, filters)
	switch err {
	case nil:
		// pass
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Unit not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load unit from database",
				"error":   err.Error(),
			})
		return
	}

	var listEntries []*UnitListEntry
	for _, unit := range units {
		entry := &UnitListEntry{
			Id:        unit.Id,
			BloodType: unit.BloodType,
			BloodRh:   unit.BloodRh,
			Status:    unit.Status,
			Location:  unit.Location,
		}
		listEntries = append(listEntries, entry)
	}

	ctx.JSON(
		http.StatusOK,
		listEntries,
	)
}

// UpdateUnit - updates the data of the specified unit
func (this *implUnitsAPI) UpdateUnit(ctx *gin.Context) {
	unitId := ctx.Param("unitId")
	if unitId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Unit ID is required",
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

	if unit.Id != "" && unitId != unit.Id {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Id missmatch (body vs query)",
			},
		)
		return
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

	existing_unit, err := db.FindDocument(ctx, unitId)
	switch err {
	case nil:
		//pass
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Unit not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to retrieve the existing unit from the database",
				"error":   err.Error(),
			},
		)
		return
	}
	unit.CreatedAt = existing_unit.CreatedAt
	unit.UpdatedAt = time.Now()
	err = db.UpdateDocument(ctx, unitId, &unit)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, unit)
		return
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Unit was deleted while processing the request",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update the unit in the database",
				"error":   err.Error(),
			},
		)
		return
	}
}

// DeleteUnit - Deletes the specific unit
func (this *implUnitsAPI) DeleteUnit(ctx *gin.Context) {
	unitId := ctx.Param("unitId")
	if unitId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Unit ID is required",
			},
		)
		return
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

	err = db.DeleteDocument(ctx, unitId)
	switch err {
	case nil:
		ctx.JSON(http.StatusNoContent, struct{}{})
		return
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Unit was deleted while processing the request",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete the unit from the database",
				"error":   err.Error(),
			},
		)
		return
	}
}
