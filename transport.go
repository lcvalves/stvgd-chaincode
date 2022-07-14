package main

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/civil"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/*
 * -----------------------------------
 * ENUMS
 * -----------------------------------
 */

type TransportationType string

const (
	Road       TransportationType = "ROAD"
	Maritime   TransportationType = "MARITIME"
	Air        TransportationType = "AIR"
	Rail       TransportationType = "RAIL"
	Intermodal TransportationType = "INTERMODAL"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Transport stores information about the transportations and shipping between operators in the supply chain
type Transport struct {
	ObjectType                  string             `json:"objType"` // objType ("t") is used to distinguish the various types of objects in state database
	ID                          string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	OriginProductionUnitID      string             `json:"originProductionUnitID"`
	DestinationProductionUnitID string             `json:"destinationProductionUnitID"`
	TransportationTypeID        TransportationType `json:"transportationTypeID"`
	InputBatch                  map[string]float32 `json:"inputBatch"`                                    // slice of single batch & quantity being shipped
	RemainingBatch              Batch              `json:"remainingBatch,omitempty" metadata:",optional"` // Only mandatory when shipped batch is partially sent (inputBatch[i].quantity < batch.quantity WHERE batch.id = i)
	Distance                    float32            `json:"distance"`                                      // in kilometers
	Cost                        float32            `json:"cost"`
	IsReturn                    bool               `json:"isReturn"`
	ActivityStartDate           civil.DateTime     `json:"activityStartDate"`
	ActivityEndDate             civil.DateTime     `json:"activityEndDate"`
	ECS                         float32            `json:"ecs"`
	SES                         float32            `json:"ses"`
}

/*
 * -----------------------------------
 * TRANSACTIONS
 * -----------------------------------
 */

// TransportnExists returns true when transport with given ID exists in world state
func (c *StvgdContract) TransportExists(ctx contractapi.TransactionContextInterface, transportID string) (bool, error) {

	// Searches for any world state data under the given transport
	data, err := ctx.GetStub().GetState(transportID)
	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateTransport creates a new instance of Transport
func (c *StvgdContract) CreateTransport(ctx contractapi.TransactionContextInterface, transportID, originProductionUnitID, destinationProductionUnitID, transportationTypeID, activityStartDate, activityEndDate string,
	inputBatch map[string]float32, distance, cost, ECS, SES float32, isReturn bool) (string, error) {

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return "", fmt.Errorf("could not read transport activity from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("transport activity [%s] already exists", transportID)
	}

	// Checks difference in ortigin & destination production IDs
	if originProductionUnitID == destinationProductionUnitID {
		return "", fmt.Errorf("origin production unit ID [%s] must be different from destination production unit ID [%s]", originProductionUnitID, destinationProductionUnitID)
	}

	// Validate transport type
	validTransportationType, err := validateTransportationType(transportationTypeID)
	if err != nil {
		return "", fmt.Errorf("could not validate transportation type. %s", err)
	}

	// Validate dates
	civilDates, err := validateDates(activityStartDate, activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates. %s", err)
	}

	// Validate scores
	validScores, err := validateScores(ECS, SES)
	if !validScores {
		return "", fmt.Errorf("invalid scores. %s", err)
	}

	// Validate distance
	if distance <= 0 {
		return "", fmt.Errorf("distance must be 0+. %s", err)
	}
	// Validate cost
	if cost <= 0 {
		return "", fmt.Errorf("cost must be 0+. %s", err)
	}

	// Input batches min length (1)
	if len(inputBatch) < 1 {
		return "", fmt.Errorf("transport must have atleast 1 input batch")
	}

	// Aux variables
	remainingBatch := new(Batch)

	for batchID, quantity := range inputBatch { // In every single input batch

		// Checks if the batch ID exists in world state
		exists, err := c.BatchExists(ctx, batchID)
		if err != nil {
			return "", fmt.Errorf("could not read batch from world state. %s", err)
		} else if !exists {
			return "", fmt.Errorf("batch [%s] does not exist", batchID)
		}

		// Reads the batch
		batch, err := c.ReadBatch(ctx, batchID)
		if err != nil {
			return "", fmt.Errorf("could not read batch from world state. %s", err)
		}

		// Validate origin production unit
		if originProductionUnitID != batch.ProductionUnitID {
			return "", fmt.Errorf("it is only possible to issue transports of batches from a production unit they belong to. %s", err)
		}

		// Validate inserted quantities (0 < quantity(inputBatch) <= batch.Quantity)
		switch {
		case quantity <= 0:
			return "", fmt.Errorf("input batch quantity must be greater than 0 (input quantity for [%s] is %.2f)", batchID, quantity)
		case quantity > batch.Quantity:
			return "", fmt.Errorf("input batch quantity must not exceed the batch's total quantity ([%s] max quantity is %.2f)", batchID, batch.Quantity)
		case quantity < batch.Quantity:

			// Initialize remaining batch object
			remainingBatch = &Batch{
				ObjectType:       batch.ObjectType,
				ID:               batch.ID + "-leftover",
				BatchTypeID:      batch.BatchTypeID,
				ProductionUnitID: batch.ProductionUnitID,
				BatchInternalID:  batch.BatchInternalID,
				SupplierID:       batch.SupplierID,
				BatchComposition: batch.BatchComposition,
				Traceability: Traceability{
					Activities:    append(batch.Traceability.Activities, transportID),
					ParentBatches: append(batch.Traceability.ParentBatches, batchID),
				},
				Quantity: batch.Quantity - quantity,
				Unit:     batch.Unit,
				ECS:      batch.ECS,
				SES:      batch.SES,
			}
			// Marshal remaining batch to bytes
			remainingBatchBytes, err := json.Marshal(remainingBatch)
			if err != nil {
				return "", err
			}
			// Put remainingBatchBytes in world state
			err = ctx.GetStub().PutState(remainingBatch.ID, remainingBatchBytes)
			if err != nil {
				return "", fmt.Errorf("failed to put batch to world state: %v", err)
			}

		default: // quantity = batch.Quantity
			// remaining batch is not created and the entire batch is shipped
		}

		// Initialize updated/"new" Batch object
		updatedInputBatch := &Batch{
			ObjectType:       batch.ObjectType,
			ID:               batch.ID,
			BatchTypeID:      batch.BatchTypeID,
			ProductionUnitID: batch.ProductionUnitID,
			BatchInternalID:  batch.BatchInternalID,
			SupplierID:       batch.SupplierID,
			BatchComposition: batch.BatchComposition,
			Traceability: Traceability{
				Activities:    append(batch.Traceability.Activities, transportID),
				ParentBatches: batch.Traceability.ParentBatches,
			},
			Quantity: quantity,
			Unit:     batch.Unit,
			ECS:      batch.ECS,
			SES:      batch.SES,
		}

		// Marshal input batch to bytes
		inputBatchBytes, err := json.Marshal(updatedInputBatch)
		if err != nil {
			return "", err
		}
		// Put inputBatchBytes in world state
		err = ctx.GetStub().PutState(updatedInputBatch.ID, inputBatchBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put batch to world state: %v", err)
		}
	}
	// Instatiate transport
	transport := &Transport{
		ObjectType:                  "t",
		ID:                          transportID,
		OriginProductionUnitID:      originProductionUnitID,
		DestinationProductionUnitID: destinationProductionUnitID,
		TransportationTypeID:        validTransportationType,
		InputBatch:                  inputBatch,
		RemainingBatch:              *remainingBatch,
		Distance:                    distance,
		Cost:                        cost,
		IsReturn:                    isReturn,
		ActivityStartDate:           civilDates[0],
		ActivityEndDate:             civilDates[1],
		ECS:                         ECS,
		SES:                         SES,
	}

	// Marshal transport to bytes
	transportBytes, err := json.Marshal(transport)
	if err != nil {
		return "", err
	}
	// Put transportBytes in world state
	err = ctx.GetStub().PutState(transportID, transportBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put transport to world state: %v", err)
	}

	return fmt.Sprintf("transport activity [%s] & batch were successfully added to the ledger", transportID), nil

}

// ReadTransport retrieves an instance of Transport from the world state
func (c *StvgdContract) ReadTransport(ctx contractapi.TransactionContextInterface, transportID string) (*Transport, error) {

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("transport [%s] does not exist", transportID)
	}

	// Queries world state for transport with given ID
	transportBytes, _ := ctx.GetStub().GetState(transportID)
	// Instatiate transport
	transport := new(Transport)
	// Unmarshal transportBytes to JSON
	err = json.Unmarshal(transportBytes, transport)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Transport")
	}

	return transport, nil
}

//! GetAllTransports returns all transports found in world state
func (c *StvgdContract) GetAllTransports(ctx contractapi.TransactionContextInterface) ([]*Transport, error) {
	// range query with empty string for endKey does an open-ended query of all transports in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("t", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var transports []*Transport
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var transport Transport
		err = json.Unmarshal(queryResponse.Value, &transport)
		if err != nil {
			return nil, err
		}
		transports = append(transports, &transport)
	}

	return transports, nil
}

// DeleteTransport deletes an instance of Transport from the world state
func (c *StvgdContract) DeleteTransport(ctx contractapi.TransactionContextInterface, transportID string) (string, error) {

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return "", fmt.Errorf("could not read transport from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("transport [%s] does not exist", transportID)
	}

	// Deletes transport in the world state
	err = ctx.GetStub().DelState(transportID)
	if err != nil {
		return "", fmt.Errorf("could not delete transport from world state. %s", err)
	} else {
		return fmt.Sprintf("transport [%s] deleted successfully", transportID), nil
	}
}

//! DeleteAllTransports deletes all transport found in world state
func (c *StvgdContract) DeleteAllTransports(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the transports in world state
	transports, err := c.GetAllTransports(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read transports from world state. %s", err)
	} else if len(transports) == 0 {
		return "", fmt.Errorf("there are no transports in world state to delete")
	}

	// Iterate through transports slice
	for _, transport := range transports {
		// Delete each transport from world state
		err = ctx.GetStub().DelState(transport.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete transports from world state. %s", err)
		}
	}

	return "all the transports were successfully deleted", nil
}
