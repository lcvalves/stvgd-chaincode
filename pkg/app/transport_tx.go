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
func (c *StvgdContract) CreateTransport(ctx contractapi.TransactionContextInterface, transportID, originProductionUnitInternalID, destinationProductionUnitID, transportTypeID, activityDate string, inputBatch map[string]float32, isReturn bool) (string, error) {

	// Activity prefix validation
	activityPrefix, err := validateActivityType(transportID)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	} else if activityPrefix != "t" {
		return "", fmt.Errorf("activity ID prefix must match its type (should be [t-...])")
	}

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return "", fmt.Errorf("could not read transport activity from world state: %w", err)
	} else if exists {
		return "", fmt.Errorf("transport activity [%s] already exists", transportID)
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

	// Validate destination production unit's ID
	if destinationProductionUnitID == "" {
		return "", fmt.Errorf("destination production unit's ID must not be empty")
	}

	// Validate transport type
	validTransportType, err := validateTransportType(transportTypeID)
	if err != nil {
		return "", fmt.Errorf("could not validate transportation type: %w", err)
	}

	// Input batches mandatory length (1)
	if len(inputBatch) != 1 {
		return "", fmt.Errorf("transport activities must only have 1 input batch")
	}

	// Aux variables
	transport := new(domain.Transport)
	auxInputBatches := map[string]domain.InputBatch{}
	remainingBatch := new(domain.Batch)
	updatedInputBatch := new(domain.Batch)

	for batchID, quantity := range inputBatch { // In every single input batch

		// Checks if the batch ID exists in world state
		exists, err := c.BatchExists(ctx, batchID)
		if err != nil {
			return "", fmt.Errorf("could not read batch from world state: %w", err)
		} else if !exists {
			return "", fmt.Errorf("batch [%s] does not exist", batchID)
		}

		// Reads the batch
		batch, err := c.ReadBatch(ctx, batchID)
		if err != nil {
			return "", fmt.Errorf("could not read batch from world state: %w", err)
		}

		// Checks difference in ortigin & destination production IDs
		if batch.LatestOwner == destinationProductionUnitID {
			return "", fmt.Errorf("origin production unit ID [%s] must be different from destination production unit ID [%s]", batch.LatestOwner, destinationProductionUnitID)
		}

		// Cannot use a batch that is in transit
		if batch.IsInTransit {
			return "", fmt.Errorf("batch [%s] currently in transit", batchID)
		}

		// Ship entire batch if it's a return transport
		if isReturn && quantity != batch.Quantity {
			return "", fmt.Errorf("when returning a batch, input batch quantity [%.2f] must be equal to batch's total quantity [%.2f]", quantity, batch.Quantity)
		}

		// Aux variables
		auxInputBatches[batchID] = domain.InputBatch{
			Batch:    batch,
			Quantity: quantity,
		}

		// Instatiate transport
		transport = &domain.Transport{
			DocType:                     "t",
			ID:                          transportID,
			OriginProductionUnitID:      mspID + ":" + originProductionUnitInternalID,
			Issuer:                      clientID,
			DestinationProductionUnitID: destinationProductionUnitID,
			TransportType:               validTransportType,
			IsReturn:                    isReturn,
			ActivityDate:                txTimestamp,
			InputBatch:                  auxInputBatches,
		}

		//TODO: validate isReturn (if isReturn=true, check if last activity is Reception w/ reception.isAccepted=false)

		// Validate inserted quantities (0 < quantity(inputBatch) <= batch.Quantity)
		switch {
		case quantity <= 0:
			return "", fmt.Errorf("input batch quantity must be greater than 0 (input quantity for [%s] is %.2f)", batchID, quantity)
		case quantity > batch.Quantity:
			return "", fmt.Errorf("input batch quantity must not exceed the batch's total quantity ([%s] max quantity is %.2f)", batchID, batch.Quantity)
		case quantity < batch.Quantity:

			// Initialize remaining batch object
			remainingBatch = &domain.Batch{
				DocType:          batch.DocType,
				ID:               batch.ID + "-leftover",
				BatchType:        batch.BatchType,
				LatestOwner:      batch.LatestOwner,
				BatchInternalID:  batch.BatchInternalID,
				SupplierID:       batch.SupplierID,
				BatchComposition: batch.BatchComposition,
				Traceability:     batch.Traceability,
				Quantity:         batch.Quantity - quantity,
				Unit:             batch.Unit,
				FinalScore:       batch.FinalScore,
			}

			// Marshal remaining batch to bytes
			remainingBatchBytes, err := json.Marshal(remainingBatch)
			if err != nil {
				return "", err
			}
			// Put remainingBatchBytes in world state
			err = ctx.GetStub().PutState(remainingBatch.ID, remainingBatchBytes)
			if err != nil {
				return "", fmt.Errorf("failed to put batch to world state: %w", err)
			}

		default: // quantity = batch.Quantity, remaining batch is not created and the entire batch is shipped
		}

		// Setup traceability
		activities := make([]interface{}, 0)
		activities = append(activities, transport)
		auxTrace := make([]interface{}, 0, 1)
		auxTrace = append(auxTrace, activities[len(activities)-1])

		// Initialize updated/"new" Batch object
		updatedInputBatch = &domain.Batch{
			DocType:          batch.DocType,
			ID:               batch.ID,
			BatchType:        batch.BatchType,
			LatestOwner:      batch.LatestOwner,
			BatchInternalID:  batch.BatchInternalID,
			SupplierID:       batch.SupplierID,
			IsInTransit:      true,
			Quantity:         quantity,
			Unit:             batch.Unit,
			FinalScore:       batch.FinalScore,
			BatchComposition: batch.BatchComposition,
			Traceability:     auxTrace,
		}

		// Marshal input batch to bytes
		inputBatchBytes, err := json.Marshal(updatedInputBatch)
		if err != nil {
			return "", err
		}
		// Put inputBatchBytes in world state
		err = ctx.GetStub().PutState(updatedInputBatch.ID, inputBatchBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put batch to world state: %w", err)
		}
	}

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

	if remainingBatch.ID != "" {
		return fmt.Sprintf("transport activity [%s] & batch [%s] were successfully added to the ledger. batch [%s] quantity was updated", transportID, remainingBatch.ID, updatedInputBatch.ID), nil
	} else {
		return fmt.Sprintf("transport activity [%s] was successfully added to the ledger", transportID), nil
	}
}

// ReadTransport retrieves an instance of Transport from the world state
func (c *StvgdContract) ReadTransport(ctx contractapi.TransactionContextInterface, transportID string) (*domain.Transport, error) {

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state: %w", err)
	} else if !exists {
		return nil, fmt.Errorf("transport [%s] does not exist", transportID)
	}

	// Queries world state for transport with given ID
	transportBytes, _ := ctx.GetStub().GetState(transportID)
	// Instatiate transport
	transport := new(domain.Transport)
	// Unmarshal transportBytes to JSON
	err = json.Unmarshal(transportBytes, transport)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Transport")
	}

	return transport, nil
}

// GetAllTransports returns all transports found in world state
func (c *StvgdContract) GetAllTransports(ctx contractapi.TransactionContextInterface) ([]*domain.Transport, error) {
	queryString := `{"selector":{"docType":"t"}}`
	return getQueryResultForQueryStringTransport(ctx, queryString)
}

// DeleteTransport deletes an instance of Transport from the world state
func (c *StvgdContract) DeleteTransport(ctx contractapi.TransactionContextInterface, transportID string) (string, error) {

	// Checks if the transport ID already exists
	exists, err := c.TransportExists(ctx, transportID)
	if err != nil {
		return "", fmt.Errorf("could not read transport from world state: %w", err)
	} else if !exists {
		return "", fmt.Errorf("transport [%s] does not exist", transportID)
	}

	// Deletes transport in the world state
	err = ctx.GetStub().DelState(transportID)
	if err != nil {
		return "", fmt.Errorf("could not delete transport from world state: %w", err)
	} else {
		return fmt.Sprintf("transport [%s] deleted successfully", transportID), nil
	}
}

// DeleteAllTransports deletes all transport found in world state
func (c *StvgdContract) DeleteAllTransports(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the transports in world state
	transports, err := c.GetAllTransports(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read transports from world state: %w", err)
	} else if len(transports) == 0 {
		return "", fmt.Errorf("there are no transports in world state to delete")
	}

	// Iterate through transports slice
	for _, transport := range transports {
		// Delete each transport from world state
		err = ctx.GetStub().DelState(transport.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete transports from world state: %w", err)
		}
	}

	return "all the transports were successfully deleted", nil
}
