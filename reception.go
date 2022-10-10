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
	DocType          string         `json:"docType"` // docType ("rc") is used to distinguish the various types of objects in state database
	ID               string         `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID string         `json:"productionUnitID"`
	IsAccepted       bool           `json:"isAccepted"`
	ActivityDate     civil.DateTime `json:"activityDate"`
	ReceivedBatch    Batch          `json:"receivedBatch"`
	ECS              float32        `json:"ecs"`
	SES              float32        `json:"ses"`
	NewBatch         Batch          `json:"newBatch,omitempty" metadata:",optional"` // Mandatory when batch is accepted (isAccepted = true)
}

/*
 * -----------------------------------
 * TRANSACTIONS / METHODS
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
func (c *StvgdContract) CreateReception(ctx contractapi.TransactionContextInterface, receptionID, productionUnitID, activityDate, receivedBatchID, newBatchID, newBatchInternalID string, isAccepted bool, ECS, SES float32) (string, error) {

	// Activity prefix validation
	activityPrefix, err := validateActivityType(receptionID)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	} else if activityPrefix != "rc" {
		return "", fmt.Errorf("activity ID prefix must match its type (should be [rc-...])")
	}

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return "", fmt.Errorf("could not read reception activity from world state: %w", err)
	} else if exists {
		return "", fmt.Errorf("reception activity [%s] already exists", receptionID)
	}

	// Reads the batch
	receivedBatch, err := c.ReadBatch(ctx, receivedBatchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state: %w", err)
	}

	// Cannot use a batch that is not in transit
	if !receivedBatch.IsInTransit {
		return "", fmt.Errorf("batch [%s] is not in transit", receivedBatch.ID)
	}

	// Validate production unit ID
	if productionUnitID == "" {
		return "", fmt.Errorf("production unit ID must not be empty")
	}

	// Checks difference in production unit ID & receivedBatch.destination production IDs
	if productionUnitID == receivedBatch.ProductionUnitID {
		return "", fmt.Errorf("production unit ID [%s] must be different from batch's production unit ID [%s]", productionUnitID, receivedBatch.ProductionUnitID)
	}

	// Validate date
	civilDate, err := civil.ParseDateTime(activityDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates: %w", err)
	}

	// Validate scores
	validScores, err := validateScores(ECS, SES)
	if !validScores {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Instatiate reception
	reception := &Reception{
		DocType:          "rc",
		ID:               receptionID,
		ProductionUnitID: productionUnitID,
		ReceivedBatch:    *receivedBatch,
		IsAccepted:       isAccepted,
		ActivityDate:     civilDate,
		ECS:              ECS,
		SES:              SES,
	}
	/*
		var transportID string

		// Iterate through latest traceability actvity
		iter := reflect.ValueOf(receivedBatch.Traceability[0]).MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			// Get field named "ID"
			if k.String() == "ID" {
				transportID = v.Elem().String()
				prefix, err := validateActivityType(transportID)
				if err != nil {
					return "", err
				} else if prefix != "t" { // Check for activity type (Transport)
					return "", fmt.Errorf("previous activity must be transport")
				}
				break
			}
		}

		transport, err := c.ReadTransport(ctx, transportID)
		if err != nil {
			return "", fmt.Errorf("could not read transport from world state: %w", err)
		}

		// Assign scores to corresponding transport
		transport.ECS = ECS
		transport.SES = SES

		// Marshal transport to bytes
		transportBytes, err := json.Marshal(transport)
		if err != nil {
			return "", err
		}

		// Put transportBytes in world state
		err = ctx.GetStub().PutState(transportID, transportBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put transport to world state: %w", err)
		}
	*/
	// Initialize updated/"new" Batch object + aux variables
	newBatch := new(Batch)
	activities := make([]interface{}, 0)
	auxTrace := make([]interface{}, 0, 1)

	if isAccepted {
		newBatch = &Batch{
			DocType:          "b",
			ID:               newBatchID,
			BatchType:        receivedBatch.BatchType,
			ProductionUnitID: productionUnitID,
			BatchInternalID:  newBatchInternalID,
			SupplierID:       receivedBatch.SupplierID,
			BatchComposition: receivedBatch.BatchComposition,
			Quantity:         receivedBatch.Quantity,
			Unit:             receivedBatch.Unit,
			ECS:              receivedBatch.ECS,
			SES:              receivedBatch.SES,
		}

		receivedBatch.Quantity = 0     // Remove quantity from "old"
		reception.NewBatch = *newBatch // Update reception with newly created batch

		// Setup new batch traceability
		activities = append(activities, reception)
		auxTrace = append(auxTrace, activities[len(activities)-1])
		newBatch.Traceability = auxTrace

		// Marshal new batch to bytes
		newBatchBytes, err := json.Marshal(newBatch)
		if err != nil {
			return "", err
		}
		// Put newBatchBytes in world state
		err = ctx.GetStub().PutState(newBatch.ID, newBatchBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put new batch to world state: %w", err)
		}
	}

	receivedBatch.IsInTransit = false // Received batch is no longer in transit

	// Marshal input batch to bytes
	receivedBatchBytes, err := json.Marshal(receivedBatch)
	if err != nil {
		return "", err
	}
	// Put receivedBatchBytes in world state
	err = ctx.GetStub().PutState(receivedBatch.ID, receivedBatchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put received batch to world state: %w", err)
	}

	// Marshal reception to bytes
	receptionBytes, err := json.Marshal(reception)
	if err != nil {
		return "", err
	}
	// Put receptionBytes in world state
	err = ctx.GetStub().PutState(reception.ID, receptionBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put reception to world state: %w", err)
	}

	// Different outputs based on new batch creation
	if newBatch.ID == "" {
		return fmt.Sprintf("reception activity [%s] was successfully added to the ledger", receptionID), nil
	} else {
		// Delete received batch when it is accepted
		err = ctx.GetStub().DelState(receivedBatch.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete batch from world state: %w", err)
		}
		return fmt.Sprintf("reception activity [%s] & batch [%s] were successfully added to the ledger. batch [%s] was deleted successfully", receptionID, newBatchID, receivedBatch.ID), nil
	}
}

// ReadReception retrieves an instance of Reception from the world state
func (c *StvgdContract) ReadReception(ctx contractapi.TransactionContextInterface, receptionID string) (*Reception, error) {

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state: %w", err)
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

// ! GetAllReceptions returns all receptions found in world state
func (c *StvgdContract) GetAllReceptions(ctx contractapi.TransactionContextInterface) ([]*Reception, error) {
	queryString := `{"selector":{"docType":"rc"}}`
	return getQueryResultForQueryStringReception(ctx, queryString)
}

// DeleteReception deletes an instance of Reception from the world state
func (c *StvgdContract) DeleteReception(ctx contractapi.TransactionContextInterface, receptionID string) (string, error) {

	// Checks if the reception ID already exists
	exists, err := c.ReceptionExists(ctx, receptionID)
	if err != nil {
		return "", fmt.Errorf("could not read reception from world state: %w", err)
	} else if !exists {
		return "", fmt.Errorf("reception [%s] does not exist", receptionID)
	}

	// Deletes reception in the world state
	err = ctx.GetStub().DelState(receptionID)
	if err != nil {
		return "", fmt.Errorf("could not delete reception from world state: %w", err)
	} else {
		return fmt.Sprintf("reception [%s] deleted successfully", receptionID), nil
	}
}

// ! DeleteAllReceptions deletes all receptions found in world state
func (c *StvgdContract) DeleteAllReceptions(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the receptions in world state
	receptions, err := c.GetAllReceptions(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read receptions from world state: %w", err)
	} else if len(receptions) == 0 {
		return "", fmt.Errorf("there are no receptions in world state to delete")
	}

	// Iterate through receptions slice
	for _, reception := range receptions {
		// Delete each reception from world state
		err = ctx.GetStub().DelState(reception.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete receptions from world state: %w", err)
		}
	}

	return "all the receptions were successfully deleted", nil
}
