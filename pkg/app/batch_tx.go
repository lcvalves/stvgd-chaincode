package app

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/lcvalves/stvgd-chaincode/pkg/domain"
)

/*
 * -----------------------------------
 * TRANSACTIONS / METHODS
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
func (c *StvgdContract) ReadBatch(ctx contractapi.TransactionContextInterface, batchID string) (*domain.Batch, error) {

	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("batch [%s] does not exist", batchID)
	}

	batchBytes, _ := ctx.GetStub().GetState(batchID)

	batch := new(domain.Batch)

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
func (c *StvgdContract) GetAvailableBatches(ctx contractapi.TransactionContextInterface) ([]*domain.Batch, error) {
	queryString := `{"selector":{"docType":"b"}}`
	return getQueryResultForQueryStringBatch(ctx, queryString)
}

/*
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

		var batch domain.Batch
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &batch)
			if err != nil {
				return nil, err
			}
		} else {
			batch = domain.Batch{
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
*/
// TraceBatchByInternalID lists the batch's traceability
func (c *StvgdContract) TraceBatchByInternalID(ctx contractapi.TransactionContextInterface, batchInternalID string) ([]*domain.Batch, error) {
	queryString := `{"selector":{"batchInternalID":"` + batchInternalID + `"}}`
	return getQueryResultForQueryStringBatch(ctx, queryString)
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

// DeleteAllBatches deletes all registrations found in world state
func (c *StvgdContract) DeleteAllBatches(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the registrations in world state
	registrations, err := c.GetAvailableBatches(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read registrations from world state. %s", err)
	} else if len(registrations) == 0 {
		return "", fmt.Errorf("there are no registrations in world state to delete")
	}

	// Iterate through registrations slice
	for _, registration := range registrations {
		// Delete each registration from world state
		err = ctx.GetStub().DelState(registration.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete registrations from world state. %s", err)
		}
	}

	return "all batches deleted successfully", nil
}
