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
	ConventionalCotton BatchType = "CONVENTIONAL_COTTON"
	OrganicCotton      BatchType = "ORGANIC_COTTON"
	RecycledCotton     BatchType = "RECYCLED_COTTON"
	Pes                BatchType = "PES"
	PesRPet            BatchType = "PES_RPET"
	Polypropylene      BatchType = "POLYPROPYLENE"
	Polyamide6         BatchType = "POLYAMIDE_6"
	Polyamide66        BatchType = "POLYAMIDE_66"
	Pan                BatchType = "PAN"
	Viscose            BatchType = "VISCOSE"
	Flax               BatchType = "FLAX"
	Jute               BatchType = "JUTE"
	Kenaf              BatchType = "KENAF"
	Bamboo             BatchType = "BAMBOO"
	Silk               BatchType = "SILK"
	Wool               BatchType = "WOOL"
	Elastane           BatchType = "ELASTANE"
	Yarn               BatchType = "YARN"
	RawFabric          BatchType = "RAW_FABRIC"
	DyedFabric         BatchType = "DYED_FABRIC"
	RawKnittedFabric   BatchType = "RAW_KNITTED_FABRIC"
	DyedKnittedFabric  BatchType = "DYED_KNITTED_FABRIC"
	Garment            BatchType = "GARMENT"
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
