package services

// import (
// 	"fmt"

// 	"ec.com/database"
// 	"ec.com/models"
// 	"github.com/google/uuid"
// )

// func CreateSoliciation() (error, models.Solicitation) {
// 	// var agency models.Agency
// 	// agencyId, err := uuid.Parse(c.Params("agency_id", ""))
// 	// if agencyId == uuid.Nil {
// 	// 	return fmt.Errorf("You need to provide a valid agency_id"), models.Solicitation{}
// 	// }

// 	if database.DB.Where("id = ?", agencyId).First(&agency).Error != nil {
// 		return fmt.Errorf("agency not found"), models.Solicitation{}
// 	}

// 	var solicitation models.Solicitation
// 	if err := c.BodyParser(&solicitation); err != nil {
// 		return fmt.Errorf("cannot parse JSON"), models.Solicitation{}
// 	}

// 	solicitation.AgencyId = agency.ID
// 	solicitation.Agency = agency.Name

// 	solicitation, err = CreateSolicitation(solicitation)
// 	if err != nil {
// 		return fmt.Errorf("agency not found"), models.Solicitation{}
// 	}

// 	return nil, solicitation
// }
