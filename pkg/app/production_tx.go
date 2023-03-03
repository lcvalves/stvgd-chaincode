package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/lcvalves/stvgd-chaincode/pkg/domain"
)

/*
 * -----------------------------------
 * TRANSACTIONS / METHODS
 * -----------------------------------
 */

// ProductionExists returns true when production with given ID exists in world state
func (c *StvgdContract) ProductionExists(ctx contractapi.TransactionContextInterface, productionID string) (bool, error) {

	// Searches for any world state data under the given production
	data, err := ctx.GetStub().GetState(productionID)
	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateProduction creates a new instance of Production
func (c *StvgdContract) CreateProduction(ctx contractapi.TransactionContextInterface, productionID, productionUnitInternalID, productionTypeID, activityStartDate, batchID, batchType, batchInternalID, supplierID, unit string, inputBatches, batchComposition map[string]float32, quantity, finalScore, productionScore, SES float32) (string, error) {

	// Activity prefix validation
	activityPrefix, err := validateActivityType(productionID)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	} else if activityPrefix != "p" {
		return "", fmt.Errorf("activity ID prefix must match its type (should be [p-...])")
	}

	// Checks if the production activity ID already exists
	exists, err := c.ProductionExists(ctx, productionID)
	if err != nil {
		return "", fmt.Errorf("could not read production activity from world state: %w", err)
	} else if exists {
		return "", fmt.Errorf("production activity [%s] already exists", productionID)
	}

	// Parse start date
	parsedStartDate, err := time.Parse(time.RFC3339, activityStartDate)
	if err != nil {
		return "", fmt.Errorf("could not parse activity start date: %w", err)
	}

	// Timestamp when the transaction was created, have the same value across all endorsers
	txTimestamp, err := getTxTimestampRFC3339Time(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("could not get transaction timestamp: %w", err)
	}

	// Checks if dates are valid
	isStartDateBeforeEndDate := parsedStartDate.Before(txTimestamp)
	if !isStartDateBeforeEndDate {
		return "", fmt.Errorf("activity start date can't be after the activity end date: %w", err)
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

	// Validate production type
	validProductionType, err := validateProductionType(productionTypeID)
	if err != nil {
		return "", fmt.Errorf("could not validate activity type: %w", err)
	}

	// Validate production score
	validProductionScore, err := validateScore(productionScore)
	if !validProductionScore {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Validate SES
	validSES, err := validateScore(SES)
	if !validSES {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Validate batch type
	validBatchType, err := validateBatchType(batchType)
	if err != nil {
		return "", fmt.Errorf("could not validate batch type: %w", err)
	}

	// Validate unir
	validUnit, err := validateUnit(unit)
	if err != nil {
		return "", fmt.Errorf("could not validate batch unit: %w", err)
	}

	/// Validate production unit internal ID
	if productionUnitInternalID == "" {
		return "", fmt.Errorf("production unit internal ID must not be empty: %w", err)
	}

	// Input batches min length (1)
	if len(inputBatches) < 1 {
		return "", fmt.Errorf("production must have atleast 1 input batch")
	}

	// Aux variables
	activities := make([]interface{}, 0)
	auxTrace := make([]interface{}, 0, 1)
	auxInputBatches := map[string]domain.InputBatch{}

	for batchID, quantity := range inputBatches { // In every single input batch

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

		// Cannot use a batch that is in transit
		if batch.IsInTransit {
			return "", fmt.Errorf("batch [%s] currently in transit", batchID)
		}

		// Validate inserted quantities (0 <= quantity(inputBatch) <= batch.Quantity)
		switch {
		case quantity <= 0:
			return "", fmt.Errorf("input batches' quantities must be greater than 0 (input quantity for [%s] is %.2f)", batchID, quantity)
		case quantity > batch.Quantity:
			return "", fmt.Errorf("input batches' quantities must not exceed the batch's total quantity ([%s] max quantity is %.2f)", batchID, batch.Quantity)
		}

		// Append input batches traceability for output batch
		activities = append(activities, batch.Traceability...)
		auxInputBatch := &domain.InputBatch{
			Batch:    batch,
			Quantity: quantity,
		}
		auxInputBatches[batchID] = *auxInputBatch

		// Initialize updated/"new" Batch object
		updatedInputBatch := &domain.Batch{
			DocType:          batch.DocType,
			ID:               batch.ID,
			BatchType:        batch.BatchType,
			LatestOwner:      batch.LatestOwner,
			BatchInternalID:  batch.BatchInternalID,
			SupplierID:       batch.SupplierID,
			Quantity:         batch.Quantity - quantity,
			Unit:             batch.Unit,
			FinalScore:       batch.FinalScore,
			BatchComposition: batch.BatchComposition,
			Traceability:     batch.Traceability,
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

		// If the entire batch is used, delete it
		if updatedInputBatch.Quantity == 0 {
			err = ctx.GetStub().DelState(updatedInputBatch.ID)
			if err != nil {
				return "", fmt.Errorf("could not delete batch from world state: %w", err)
			}
		}
	}

	// Initialize "new" Batch object
	outputBatch := &domain.Batch{
		DocType:          "b",
		ID:               batchID,
		BatchType:        validBatchType,
		LatestOwner:      mspID + ":" + productionUnitInternalID,
		BatchInternalID:  batchInternalID,
		SupplierID:       supplierID,
		Quantity:         quantity,
		Unit:             validUnit,
		FinalScore:       finalScore,
		BatchComposition: batchComposition,
	}

	// Validate output batch
	isValidBatch, err := validateBatch(ctx, outputBatch.ID, outputBatch.LatestOwner, outputBatch.BatchInternalID, outputBatch.SupplierID, string(outputBatch.Unit), string(outputBatch.BatchType), outputBatch.BatchComposition, outputBatch.Quantity, outputBatch.FinalScore, outputBatch.IsInTransit)
	if !isValidBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %w", err)
	}

	// Instantiate production
	production := &domain.Production{
		DocType:           "p",
		ID:                productionID,
		ProductionUnitID:  mspID + ":" + productionUnitInternalID,
		Issuer:            clientID,
		ProductionType:    validProductionType,
		ActivityStartDate: parsedStartDate,
		ActivityEndDate:   txTimestamp,
		ProductionScore:   productionScore,
		SES:               SES,
		InputBatches:      auxInputBatches,
		OutputBatch:       *outputBatch,
	}

	// Setup & append traceability to output batch
	activities = append(activities, production)
	auxTrace = append(auxTrace, activities[len(activities)-1])
	outputBatch.Traceability = auxTrace

	// Marshal batch to bytes
	batchBytes, err := json.Marshal(outputBatch)
	if err != nil {
		return "", err
	}
	// Put batchBytes in world state
	err = ctx.GetStub().PutState(outputBatch.ID, batchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put batch to world state: %w", err)
	}

	// Marshal production to bytes
	productionBytes, err := json.Marshal(production)
	if err != nil {
		return "", err
	}
	// Put productionBytes in world state
	err = ctx.GetStub().PutState(productionID, productionBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put production to world state: %w", err)
	}

	return fmt.Sprintf("production activity [%s] & batch [%s] were successfully added to the ledger", productionID, outputBatch.ID), nil
}

// ReadProduction retrieves an instance of Production from the world state
func (c *StvgdContract) ReadProduction(ctx contractapi.TransactionContextInterface, productionID string) (*domain.Production, error) {

	// Checks if the production ID already exists
	exists, err := c.ProductionExists(ctx, productionID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the production activity [%s] does not exist", productionID)
	}

	// Queries world state for production with given ID
	productionBytes, _ := ctx.GetStub().GetState(productionID)
	// Instatiate production
	production := new(domain.Production)
	// Unmarshal productionBytes to JSON
	err = json.Unmarshal(productionBytes, production)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Production")
	}

	return production, nil
}

// ! GetAllProductions returns all productions found in world state
func (c *StvgdContract) GetAllProductions(ctx contractapi.TransactionContextInterface) ([]*domain.Production, error) {
	queryString := `{"selector":{"docType":"p"}}`
	return getQueryResultForQueryStringProduction(ctx, queryString)
}

// DeleteProduction deletes an instance of Production from the world state
func (c *StvgdContract) DeleteProduction(ctx contractapi.TransactionContextInterface, productionID string) (string, error) {

	// Checks if the production ID already exists
	exists, err := c.ProductionExists(ctx, productionID)
	if err != nil {
		return "", fmt.Errorf("could not read production from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("production [%s] does not exist", productionID)
	}

	// Deletes production in the world state
	err = ctx.GetStub().DelState(productionID)
	if err != nil {
		return "", fmt.Errorf("could not delete production from world state. %s", err)
	} else {
		return fmt.Sprintf("production [%s] deleted successfully", productionID), nil
	}
}

// ! DeleteAllProductions deletes all production found in world state
func (c *StvgdContract) DeleteAllProductions(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the productions in world state
	productions, err := c.GetAllProductions(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read productions from world state. %s", err)
	} else if len(productions) == 0 {
		return "", fmt.Errorf("there are no productions in world state to delete")
	}

	// Iterate through productions slice
	for _, production := range productions {
		// Delete each production from world state
		err = ctx.GetStub().DelState(production.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete productions from world state. %s", err)
		}
	}

	return "all the productions were successfully deleted", nil
}
