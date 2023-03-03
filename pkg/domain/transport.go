package domain

import (
	"time"
)

/*
 * -----------------------------------
 * ENUMS
 * -----------------------------------
 */

type TransportType string

const (
	TerrestrialSmall TransportType = "TERRESTRIAL_SMALL"
	TerrestrialBig   TransportType = "TERRESTRIAL_BIG"
	Maritime         TransportType = "MARITIME"
	Aerial           TransportType = "AERIAL"
	Railroader       TransportType = "RAILROADER"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Transport stores information about the transportations and shipping between operators in the supply chain
type Transport struct {
	DocType                     string                `json:"docType"`                     // docType ("t") is used to distinguish the various types of objects in state database
	ID                          string                `json:"ID"`                          // the field tags are needed to keep case from bouncing around
	OriginProductionUnitID      string                `json:"originProductionUnitID"`      // where the tx was issued (format: <CompanyID>:<ProductionUnitIntenalID>)
	Issuer                      string                `json:"issuer"`                      // client ID of who issued the tx
	DestinationProductionUnitID string                `json:"destinationProductionUnitID"` // next owner
	TransportType               TransportType         `json:"transportationType"`          // production type enum
	IsReturn                    bool                  `json:"isReturn"`                    // when reception is not accepted, must return batch
	ActivityDate                time.Time             `json:"activityDate"`                // start of transport (end of transport is the corresponding Reception's activity date)
	InputBatch                  map[string]InputBatch `json:"inputBatch"`                  // slice of single batch & quantity being shipped
}
