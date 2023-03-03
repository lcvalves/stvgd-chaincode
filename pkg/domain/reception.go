package domain

import (
	"time"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Reception stores information about the batch receptions in the supply chain companies/production units
type Reception struct {
	DocType          string    `json:"docType"`                                 // docType ("rc") is used to distinguish the various types of objects in state database
	ID               string    `json:"ID"`                                      // the field tags are needed to keep case from bouncing around
	ProductionUnitID string    `json:"productionUnitID"`                        // where the tx was issued (format: <CompanyID>:<ProductionUnitIntenalID>)
	Issuer           string    `json:"issuer"`                                  // client ID of who issued the tx
	ActivityDate     time.Time `json:"activityDate"`                            // end of transport (start of transport is the corresponding Transport's activity date)
	IsAccepted       bool      `json:"isAccepted"`                              // quality assessment
	TransportScore   float32   `json:"transportScore"`                          // score of transport activity
	SES              float32   `json:"ses"`                                     // Social-Economic Score (yearly audited by outsider cerification entity)
	Distance         float32   `json:"distance"`                                // in kilometers                          // transport cost
	ReceivedBatch    Batch     `json:"receivedBatch"`                           // batch in reception
	NewBatch         Batch     `json:"newBatch,omitempty" metadata:",optional"` // Mandatory when batch is accepted (isAccepted = true)
}
