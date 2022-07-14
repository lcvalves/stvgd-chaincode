package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

type Traceability struct {
	Activities    []string `json:"activities"`
	ParentBatches []string `json:"parentBatches"`
}

// Batch stores information about the batches in the supply chain
type Batch struct {
	ObjectType       string             `json:"objType"` // docType ("b") is used to distinguish the various types of objects in state database
	ID               string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	BatchTypeID      BatchType          `json:"batchTypeID"`
	ProductionUnitID string             `json:"productionUnitID"` // Current owner
	BatchInternalID  string             `json:"batchInternalID"`
	SupplierID       string             `json:"supplierID"`
	BatchComposition map[string]float32 `json:"batchCompostion"` // i.e. {raw_material_id: %}
	Traceability     Traceability       `json:"traceability,omitempty" metadata:",optional"`
	Quantity         float32            `json:"quantity"`
	Unit             Unit               `json:"unit"`
	ECS              float32            `json:"ecs"`
	SES              float32            `json:"ses"`
}

/*
 * -----------------------------------
 * TRANSACTIONS
 * -----------------------------------
 */

// BatchExists returns true when batch with given ID exists in world state
func (c *StvgdContract) BatchExists(ctx contractapi.TransactionContextInterface, batchID string) (bool, error) {
	data, err := ctx.GetStub().GetState(batchID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// ReadBatch retrieves an instance of Batch from the world state
func (c *StvgdContract) ReadBatch(ctx contractapi.TransactionContextInterface, batchID string) (*Batch, error) {

	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("[%s] does not exist", batchID)
	}

	batchBytes, _ := ctx.GetStub().GetState(batchID)

	batch := new(Batch)

	err = json.Unmarshal(batchBytes, batch)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Batch")
	}

	return batch, nil
}

// GetAllBatches queries for all batches.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (docType).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (c *StvgdContract) GetAllBatches(ctx contractapi.TransactionContextInterface) ([]*Batch, error) {
	queryString := `{"selector":{"objType":"b"}}`
	return getQueryResultForQueryStringBatch(ctx, queryString)
}

// GetAssetHistory returns the chain of custody for a batch since issuance.
func (c *StvgdContract) GetBatchHistory(ctx contractapi.TransactionContextInterface, batchID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", batchID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(batchID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var batch Batch
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &batch)
			if err != nil {
				return nil, err
			}
		} else {
			batch = Batch{
				ID: batchID,
			}
		}

		timestamp := timestamppb.New(response.Timestamp.AsTime())
		if timestamp.CheckValid() != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp.AsTime(),
			Record:    &batch,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// DeleteBatch deletes an instance of Batch from the world state
func (c *StvgdContract) DeleteBatch(ctx contractapi.TransactionContextInterface, batchID string) (string, error) {
	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("[%s] does not exist", batchID)
	}

	err = ctx.GetStub().DelState(batchID)
	if err != nil {
		return "", fmt.Errorf("could not delete batch from world state. %s", err)
	} else {
		return fmt.Sprintf("[%s] deleted successfully", batchID), nil
	}
}
