package sprava_krvi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Marek-FIIT/sprava-krvi-webapi/internal/db_service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (this *implDonorsAPI) GetDonors(ctx *gin.Context) {
	// ctx.AbortWithStatus(http.StatusNotImplemented)
	filters := make(map[string]interface{})
	if bloodType := ctx.Query("bloodType"); bloodType != "" {
		filters["bloodtype"] = bloodType
	}
	if bloodRh := ctx.Query("bloodRh"); bloodRh != "" {
		filters["bloodrh"] = bloodRh
	}
	if eligible := ctx.Query("eligible"); eligible != "" {
		eligibleBool, err := strconv.ParseBool(eligible)
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
		filters["eligible"] = eligibleBool
	}

	// log.Printf("filters: %v", filters)

	db, err := db_service.GetDbService[Donor](ctx, "db_service_donors")
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

	donors, err := db.FindDocuments(ctx, filters)
	switch err {
	case nil:
		// pass
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Donor not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load donor from database",
				"error":   err.Error(),
			})
		return
	}

	var listEntries []*DonorListEntry
	for _, donor := range donors {
		entry := &DonorListEntry{
			Id:           donor.Id,
			FirstName:    donor.FirstName,
			LastName:     donor.LastName,
			BloodType:    donor.BloodType,
			BloodRh:      donor.BloodRh,
			Eligible:     donor.Eligible,
			LastDonation: donor.LastDonation,
		}
		listEntries = append(listEntries, entry)
	}

	ctx.JSON(
		http.StatusOK,
		listEntries,
	)
}

func (this *implDonorsAPI) GetDonor(ctx *gin.Context) {
	// ctx.AbortWithStatus(http.StatusNotImplemented)
	donorId := ctx.Param("donorId")
	if donorId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Donor ID is required",
			},
		)
		return
	}

	db, err := db_service.GetDbService[Donor](ctx, "db_service_donors")
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

	donor, err := db.FindDocument(ctx, donorId)

	switch err {
	case nil:
		ctx.JSON(
			http.StatusOK,
			donor,
		)
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Donor not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load donor from database",
				"error":   err.Error(),
			})
		return
	}
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

	db, err := db_service.GetDbService[Donor](ctx, "db_service_donors")
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
	donorId := ctx.Param("donorId")
	if donorId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Donor ID is required",
			},
		)
		return
	}

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

	if donor.Id != "" && donorId != donor.Id {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Id missmatch (body vs query)",
			},
		)
		return
	}

	db, err := db_service.GetDbService[Donor](ctx, "db_service_donors")
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

	existing_donor, err := db.FindDocument(ctx, donorId)
	switch err {
	case nil:
		//pass
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Donor not found",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to retrieve the existing donor from the database",
				"error":   err.Error(),
			},
		)
		return
	}
	donor.Id = existing_donor.Id
	donor.CreatedAt = existing_donor.CreatedAt
	donor.UpdatedAt = time.Now()
	err = db.UpdateDocument(ctx, donorId, &donor)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, donor)
		return
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Donor was deleted while processing the request",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update the donor in the database",
				"error":   err.Error(),
			},
		)
		return
	}
}
func (this *implDonorsAPI) DeleteDonor(ctx *gin.Context) {
	donorId := ctx.Param("donorId")
	if donorId == "" {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"status":  http.StatusBadRequest,
				"message": "Donor ID is required",
			},
		)
		return
	}

	db, err := db_service.GetDbService[Donor](ctx, "db_service_donors")
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

	err = db.DeleteDocument(ctx, donorId)
	switch err {
	case nil:
		ctx.JSON(http.StatusNoContent, struct{}{})
		return
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Donor was deleted while processing the request",
				"error":   err.Error(),
			},
		)
		return
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete the donor from the database",
				"error":   err.Error(),
			},
		)
		return
	}
}
