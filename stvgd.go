/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import "cloud.google.com/go/civil"

//* CURRENT MODEL VERSION: on-chain_struct_model_v9

// ENUMS
/* type Unit string

const (
	Kilograms     Unit = "KG"
	Liters        Unit = "L"
	Meters        Unit = "M"
	SquaredMeters Unit = "M^2"
)

type ActivityType string

const (
	Spinning        ActivityType = "SPINNING"
	Weaving         ActivityType = "WEAVING"
	Knitting        ActivityType = "KNITTING"
	DyeingFinishing ActivityType = "DYEING_FINISHING"
	Confection      ActivityType = "CONFECTION"
)

type TransportationType string

const (
	Road       TransportationType = "ROAD"
	Maritime   TransportationType = "MARITIME"
	Air        TransportationType = "AIR"
	Rail       TransportationType = "RAIL"
	Intermodal TransportationType = "INTERMODAL"
)

type BatchType string

const (
	Fiber          BatchType = "FIBER"
	Yarn           BatchType = "YARN"
	Mesh           BatchType = "MESH"
	Fabric         BatchType = "FABRIC"
	DyedMesh       BatchType = "DYED_MESH"
	FinishedMesh   BatchType = "FINISHED_MESH"
	DyedFabric     BatchType = "DYED_FABRIC"
	FinishedFabric BatchType = "FINISHED_FABRIC"
	Cut            BatchType = "CUT"
	FinishedPiece  BatchType = "FINISHED_PIECE"
)
*/

// Batch stores information about the batches in the supply chain
type Batch struct {
	ObjectType           string             `json:"docType"`                                             // docType ("batch") is used to distinguish the various types of objects in state database
	ID                   string             `json:"ID"`                                                  // the field tags are needed to keep case from bouncing around
	BatchTypeID          string             `json:"batchTypeID"`                                         //? Convert to enum
	ProductionActivityID string             `json:"productionActivityID,omitempty" metadata:",optional"` // Only mandatory when batch is produced within the system's supply chain
	ProductionUnitID     string             `json:"productionUnitID"`                                    // Current owner
	BatchInternalID      string             `json:"batchInternalID"`
	SupplierID           string             `json:"supplierID"`
	BatchComposition     map[string]float32 `json:"batchCompostion"` // i.e. {raw_material_id: %}
	Quantity             float32            `json:"quantity"`
	Unit                 string             `json:"unit"`
	ECS                  float32            `json:"ecs"`
	SES                  float32            `json:"ses"`
}

// ProductionActivity stores information about the production activities in the supply chain
type ProductionActivity struct {
	ObjectType        string             `json:"docType"` // docType ("pa") is used to distinguish the various types of objects in state database
	ID                string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string             `json:"productionUnitID"`
	CompanyID         string             `json:"companyID"`
	ActivityTypeID    string             `json:"activityTypeID"`
	InputBatches      map[string]float32 `json:"inputBatches,omitempty" metadata:",optional"` // inputLots nullable for ProdActivities that create new lots as raw material not sourced from outside the system
	OutputBatch       Batch              `json:"outputBatch"`
	ActivityStartDate civil.DateTime     `json:"activityStartDate"`
	ActivityEndDate   civil.DateTime     `json:"activityEndDate"`
	ECS               float32            `json:"ecs"`
	SES               float32            `json:"ses"`
}

// LogisticalActivityTransport stores information about the transportations and shipping between operators in the supply chain
type LogisticalActivityTransport struct {
	ObjectType                  string             `json:"docType"` // docType ("lat") is used to distinguish the various types of objects in state database
	ID                          string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	OriginProductionUnitID      string             `json:"OriginProductionUnitID"`
	DestinationProductionUnitID string             `json:"DestinationProductionUnitID"`
	TransportationTypeID        string             `json:"transportationTypeID"`
	InputBatch                  map[string]float32 `json:"inputBatches"`                                  // slice of lots being shipped
	RemainingBatch              Batch              `json:"remainingBatch,omitempty" metadata:",optional"` // Only mandatory when shipped batch is partially sent (inputBatch[i].quantity < batch.quantity WHERE batch.id = i)
	Distance                    float32            `json:"distance"`                                      // in kilometers
	Cost                        float32            `json:"cost"`
	IsReturn                    bool               `json:"isReturn"` // default = true (defined on init/CreateLogisiticalActivityTransport)
	ActivityStartDate           civil.DateTime     `json:"activityStartDate"`
	ActivityEndDate             civil.DateTime     `json:"activityEndDate"`
	ECS                         float32            `json:"ecs"`
	SES                         float32            `json:"ses"`
}

// LogisticalActivityReception stores information about the batch receptions in the supply chain companies/production units
type LogisticalActivityReception struct {
	ObjectType         string         `json:"docType"` // docType ("lar") is used to distinguish the various types of objects in state database
	ID                 string         `json:"ID"`      // the field tags are needed to keep case from bouncing around
	NewBatchInternalID string         `json:"newBatchInternalID"`
	ReceivedBatch      Batch          `json:"receivedBatch"` // Only mandatory when shipped batch is partially sent (inputBatch[i].quantity < batch.quantity WHERE batch.id = i)
	IsAccepted         bool           `json:"acceptedQuantity"`
	ActivityStartDate  civil.DateTime `json:"activityStartDate"`
	ActivityEndDate    civil.DateTime `json:"activityEndDate"`
}
