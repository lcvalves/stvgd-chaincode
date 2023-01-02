package domain

/*
 * -----------------------------------
 * ENUMS
 * -----------------------------------
 */

type Unit string

const (
	Kilograms     Unit = "KG"
	Liters        Unit = "L"
	Meters        Unit = "M"
	SquaredMeters Unit = "M2"
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
	Other          BatchType = "OTHER"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Batch stores information about the batches in the supply chain
type Batch struct {
	DocType          string             `json:"docType"` // docType ("b") is used to distinguish the various types of objects in state database
	ID               string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	BatchType        BatchType          `json:"batchType"`
	LatestOwner      string             `json:"latestOwner"` // current/latest owner
	BatchInternalID  string             `json:"batchInternalID"`
	SupplierID       string             `json:"supplierID"`
	IsInTransit      bool               `json:"isInTransit" metadata:",optional"`
	Quantity         float32            `json:"quantity"`
	Unit             Unit               `json:"unit"`
	FinalScore       float32            `json:"finalScore"`
	BatchComposition map[string]float32 `json:"batchComposition"` // i.e. {raw_material_id: %}
	Traceability     []interface{}      `json:"traceability,omitempty" metadata:",optional"`
}

type InputBatch struct {
	Batch    *Batch  `json:"batch"`
	Quantity float32 `json:"quantity"`
}
