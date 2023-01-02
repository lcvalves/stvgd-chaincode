package domain

import (
	"time"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Registration stores information about the batch registrations in the supply chain companies/production units
type Registration struct {
	DocType          string    `json:"docType"`          // docType ("rg") is used to distinguish the various types of objects in state database
	ID               string    `json:"ID"`               // the field tags are needed to keep case from bouncing around
	ProductionUnitID string    `json:"productionUnitID"` // where the tx was issued (format: <CompanyID>:<ProductionUnitIntenalID>)
	Issuer           string    `json:"issuer"`           // client ID of who issued the tx
	ActivityDate     time.Time `json:"activityDate"`     // tx timestamp
	NewBatch         Batch     `json:"newBatch"`         // newly registered batch
}
