package domain

import (
	"time"
)

/*
 * -----------------------------------
 * ENUMS
 * -----------------------------------
 */

type ProductionType string

const (
	Spinning        ProductionType = "SPINNING"
	Weaving         ProductionType = "WEAVING"
	Knitting        ProductionType = "KNITTING"
	DyeingFinishing ProductionType = "DYEING_FINISHING"
	Confection      ProductionType = "CONFECTION"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Production stores information about the production activities in the supply chain
type Production struct {
	DocType           string                `json:"docType"`           // docType ("p") is used to distinguish the various types of objects in state database
	ID                string                `json:"ID"`                // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string                `json:"productionUnitID"`  // where the tx was issued (format: <CompanyID>:<ProductionUnitIntenalID>)
	Issuer            string                `json:"issuer"`            // client ID of who issued the tx
	ProductionType    ProductionType        `json:"productionType"`    // production type enum
	ActivityStartDate time.Time             `json:"activityStartDate"` // client static data waiting on production finish
	ActivityEndDate   time.Time             `json:"activityEndDate"`   // tx timestamp
	ProductionScore   float32               `json:"productionScore"`   // score of production activity
	SES               float32               `json:"ses"`               // Social-Economic Score (yearly audited by outsider cerification entity)
	OutputBatch       Batch                 `json:"outputBatch"`       // produced batch
	InputBatches      map[string]InputBatch `json:"inputBatches"`      // batches & respective quantity to be used in production
}
