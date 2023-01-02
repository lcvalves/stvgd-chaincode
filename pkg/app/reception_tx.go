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
func (c *StvgdContract) CreateReception(ctx contractapi.TransactionContextInterface, receptionID, productionUnitInternalID, activityDate, receivedBatchID, newBatchID, newBatchInternalID string, isAccepted bool, transportScore, SES, distance, cost float32) (string, error) {

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

	// Timestamp when the transaction was created, have the same value across all endorsers
	txTimestamp, err := getTxTimestampRFC3339Time(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("could not get transaction timestamp: %w", err)
	}

	// Get company MSP ID
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not get MSP ID: %w", err)
	}

	// Get issuer client ID
	clientID, err := getSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get issuer's client ID: %w", err)
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
	if productionUnitInternalID == "" {
		return "", fmt.Errorf("production unit internal ID must not be empty")
	}

	// production unit ID composite key
	productionUnitID := mspID + ":" + productionUnitInternalID

	_ = iterate(receivedBatch.Traceability)
	/*
		//TODO: check last transport destination
		// ValueOf(transporID)
		transportDestination := iterate(receivedBatch.Traceability)

		if true {
			return fmt.Sprintf("%v", transportDestination), nil
		}

			// Read transport
			transport, err := c.ReadTransport(ctx, reflect.ValueOf(transportID).String())
			if err != nil {
				return "", fmt.Errorf("could not read transport from world state: %w", err)
			}

				//TODO: Checks if reception's produciton unit is the destination of the corresponding transport
				if transport.DestinationProductionUnitID != productionUnitID {
					return "", fmt.Errorf("received batch [%s] destination [%s] is not equal to the issuer's production unit [%s]", receivedBatch.ID, transport, transport.DestinationProductionUnitID, productionUnitID)
				}
	*/
	// Checks difference in production unit ID & receivedBatch.destination production IDs
	// Avoids transport to self production unit
	if productionUnitID == receivedBatch.LatestOwner {
		return "", fmt.Errorf("production unit ID [%s] must be different from batch's production unit ID [%s]", productionUnitID, receivedBatch.LatestOwner)
	}

	// Validate transport score
	validTransportScore, err := validateScore(transportScore)
	if !validTransportScore {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Validate SES
	validSES, err := validateScore(SES)
	if !validSES {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Validate distance
	if distance <= 0 {
		return "", fmt.Errorf("distance must be 0+")
	}
	// Validate cost
	if cost <= 0 {
		return "", fmt.Errorf("cost must be 0+")
	}

	// Instatiate reception
	reception := &domain.Reception{
		DocType:          "rc",
		ID:               receptionID,
		ProductionUnitID: productionUnitID,
		Issuer:           clientID,
		ActivityDate:     txTimestamp,
		ReceivedBatch:    *receivedBatch,
		IsAccepted:       isAccepted,
		TransportScore:   transportScore,
		SES:              SES,
		Distance:         distance,
		Cost:             cost,
	}

	// Initialize updated/"new" Batch object + aux variables
	newBatch := new(domain.Batch)
	activities := make([]interface{}, 0)
	auxTrace := make([]interface{}, 0, 1)

	if isAccepted {
		newBatch = &domain.Batch{
			DocType:          "b",
			ID:               newBatchID,
			BatchType:        receivedBatch.BatchType,
			LatestOwner:      productionUnitID,
			BatchInternalID:  newBatchInternalID,
			SupplierID:       receivedBatch.SupplierID,
			BatchComposition: receivedBatch.BatchComposition,
			Quantity:         receivedBatch.Quantity,
			Unit:             receivedBatch.Unit,
			FinalScore:       receivedBatch.FinalScore,
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

	// Received batch is no longer in transit
	receivedBatch.IsInTransit = false

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

	//? IS THIS CORRECT?
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
func (c *StvgdContract) ReadReception(ctx contractapi.TransactionContextInterface, receptionID string) (*domain.Reception, error) {

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
	reception := new(domain.Reception)
	// Unmarshal receptionBytes to JSON
	err = json.Unmarshal(receptionBytes, reception)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Reception")
	}

	return reception, nil
}

// ! GetAllReceptions returns all receptions found in world state
func (c *StvgdContract) GetAllReceptions(ctx contractapi.TransactionContextInterface) ([]*domain.Reception, error) {
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
