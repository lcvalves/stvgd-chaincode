package main

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/civil"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Reception stores information about the batch receptions in the supply chain companies/production units
type Reception struct {
	ObjectType        string         `json:"docType"` // docType ("rc") is used to distinguish the various types of objects in state database
	ID                string         `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string         `json:"productionUnitID"`
	ReceivedBatch     Batch          `json:"receivedBatch"`
	NewBatch          Batch          `json:"newBatch,omitempty" metadata:",optional"` // Mandatory when batch is accepted (isAccepted = true)
	IsAccepted        bool           `json:"isAccepted"`
	ActivityStartDate civil.DateTime `json:"activityStartDate"`
	ActivityEndDate   civil.DateTime `json:"activityEndDate"`
}

/*
 * -----------------------------------
 * TRANSACTIONS
 * -----------------------------------
 */

// ReceptionnExists returns true when reception with given ID exists in world state
func (c *StvgdContract) ReceptionExists(ctx contractapi.TransactionContextInterface, receptionID string) (bool, error) {

	// Searches for any world state data under the given reception
	data, err := ctx.GetStub().GetState(receptionID)
	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateReception creates a new instance of Reception
func (c *StvgdContract) CreateReception(ctx contractapi.TransactionContextInterface, receptionID, productionUnitID,
	activityStartDate, activityEndDate, receivedBatchID, newBatchID, newBatchInternalID string, isAccepted bool) (string, error) {

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return "", fmt.Errorf("could not read reception activity from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("reception activity [%s] already exists", receptionID)
	}

	// Reads the batch
	receivedBatch, err := c.ReadBatch(ctx, receivedBatchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	}

	// Checks difference in production unit ID & receivedBatch.destination production IDs
	if productionUnitID == receivedBatch.ProductionUnitID {
		return "", fmt.Errorf("production unit ID [%s] must be different from batch's production unit ID [%s]", productionUnitID, receivedBatch.ProductionUnitID)
	}

	// Validate dates
	civilDates, err := validateDates(activityStartDate, activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates. %s", err)
	}
	newBatch := new(Batch)
	if isAccepted {

		// Initialize updated/"new" Batch object
		newBatch = &Batch{
			ObjectType:       "b",
			ID:               newBatchID,
			BatchTypeID:      receivedBatch.BatchTypeID,
			ProductionUnitID: productionUnitID,
			BatchInternalID:  newBatchInternalID,
			SupplierID:       receivedBatch.SupplierID,
			BatchComposition: receivedBatch.BatchComposition,
			Traceability: Traceability{
				Activities:    append(receivedBatch.Traceability.Activities, receptionID),
				ParentBatches: append(receivedBatch.Traceability.ParentBatches, receivedBatch.ID),
			},
			Quantity: receivedBatch.Quantity,
			Unit:     receivedBatch.Unit,
			ECS:      receivedBatch.ECS,
			SES:      receivedBatch.SES,
		}

		// Marshal input batch to bytes
		newBatchBytes, err := json.Marshal(newBatch)
		if err != nil {
			return "", err
		}
		// Put newBatchBytes in world state
		err = ctx.GetStub().PutState(newBatch.ID, newBatchBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put batch to world state: %v", err)
		}

	}
	// Instatiate reception
	reception := &Reception{
		ObjectType:        "rc",
		ID:                receptionID,
		ProductionUnitID:  productionUnitID,
		ReceivedBatch:     *receivedBatch,
		NewBatch:          *newBatch,
		IsAccepted:        isAccepted,
		ActivityStartDate: civilDates[0],
		ActivityEndDate:   civilDates[1],
	}

	// Marshal reception to bytes
	receptionBytes, err := json.Marshal(reception)
	if err != nil {
		return "", err
	}
	// Put receptionBytes in world state
	err = ctx.GetStub().PutState(receptionID, receptionBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put reception to world state: %v", err)
	}

	return fmt.Sprintf("reception activity [%s] & batch [%s] were successfully added to the ledger", receptionID, newBatchID), nil

}

// ReadReception retrieves an instance of Reception from the world state
func (c *StvgdContract) ReadReception(ctx contractapi.TransactionContextInterface, receptionID string) (*Reception, error) {

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("reception [%s] does not exist", receptionID)
	}

	// Queries world state for reception with given ID
	receptionBytes, _ := ctx.GetStub().GetState(receptionID)
	// Instatiate reception
	reception := new(Reception)
	// Unmarshal receptionBytes to JSON
	err = json.Unmarshal(receptionBytes, reception)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Reception")
	}

	return reception, nil
}

// GetAllReceptions returns all receptions found in world state
func (c *StvgdContract) GetAllReceptions(ctx contractapi.TransactionContextInterface) ([]*Reception, error) {
	// range query with empty string for endKey does an open-ended query of all receptions in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("rg", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var receptions []*Reception
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var reception Reception
		err = json.Unmarshal(queryResponse.Value, &reception)
		if err != nil {
			return nil, err
		}
		receptions = append(receptions, &reception)
	}

	return receptions, nil
}

// DeleteReception deletes an instance of Reception from the world state
func (c *StvgdContract) DeleteReception(ctx contractapi.TransactionContextInterface, receptionID string) (string, error) {

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return "", fmt.Errorf("could not read reception from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("reception [%s] does not exist", receptionID)
	}

	// Deletes reception in the world state
	err = ctx.GetStub().DelState(receptionID)
	if err != nil {
		return "", fmt.Errorf("could not delete reception from world state. %s", err)
	} else {
		return fmt.Sprintf("reception [%s] deleted successfully", receptionID), nil
	}
}

// DeleteAllReceptions deletes all receptions found in world state
func (c *StvgdContract) DeleteAllReceptions(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the receptions in world state
	receptions, err := c.GetAllReceptions(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read receptions from world state. %s", err)
	} else if len(receptions) == 0 {
		return "", fmt.Errorf("there are no receptions in world state to delete")
	}

	// Iterate through receptions slice
	for _, reception := range receptions {
		// Delete each reception from world state
		err = ctx.GetStub().DelState(reception.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete receptions from world state. %s", err)
		}
	}

	return "all the receptions were successfully deleted", nil
}
