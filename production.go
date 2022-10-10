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

type ProductionType string

const (
	Spinning        ProductionType = "SPINNING"
	Weaving         ProductionType = "WEAVING"
	Knitting        ProductionType = "KNITTING"
	DyeingFinishing ProductionType = "DYEING_FINISHING"
	Confection      ProductionType = "CONFECTION"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Production stores information about the production activities in the supply chain
type Production struct {
	DocType           string                `json:"docType"` // docType ("p") is used to distinguish the various types of objects in state database
	ID                string                `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string                `json:"productionUnitID"`
	CompanyID         string                `json:"companyID"`
	ProductionType    ProductionType        `json:"productionType"`
	ActivityStartDate civil.DateTime        `json:"activityStartDate"`
	ActivityEndDate   civil.DateTime        `json:"activityEndDate"`
	ECS               float32               `json:"ecs"`
	SES               float32               `json:"ses"`
	OutputBatch       Batch                 `json:"outputBatch"`
	InputBatches      map[string]InputBatch `json:"inputBatches"`
}

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
func (c *StvgdContract) CreateProduction(ctx contractapi.TransactionContextInterface, productionID, companyID, productionTypeID, activityStartDate, activityEndDate string, inputBatches map[string]float32, outputBatch Batch, ECS, SES float32) (string, error) {

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

	// Validate company ID
	if companyID == "" {
		return "", fmt.Errorf("company ID must not be empty")
	}

	// Validate production type
	validProductionType, err := validateProductionType(productionTypeID)
	if err != nil {
		return "", fmt.Errorf("could not validate activity type: %w", err)
	}

	// Validate dates
	civilDates, err := validateDates(activityStartDate, activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates: %w", err)
	}

	// Validate scores
	validScores, err := validateScores(ECS, SES)
	if !validScores {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Input batches min length (1)
	if len(inputBatches) < 1 {
		return "", fmt.Errorf("production must have atleast 1 input batch")
	}

	// Aux variables
	activities := make([]interface{}, 0)
	auxTrace := make([]interface{}, 0, 1)
	auxInputBatches := map[string]InputBatch{}

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
		auxInputBatch := InputBatch{
			Batch:    batch,
			Quantity: quantity,
		}
		auxInputBatches[batchID] = auxInputBatch

		// Initialize updated/"new" Batch object
		updatedInputBatch := &Batch{
			DocType:          batch.DocType,
			ID:               batch.ID,
			BatchType:        batch.BatchType,
			ProductionUnitID: batch.ProductionUnitID,
			BatchInternalID:  batch.BatchInternalID,
			SupplierID:       batch.SupplierID,
			Quantity:         batch.Quantity - quantity,
			Unit:             batch.Unit,
			ECS:              batch.ECS,
			SES:              batch.SES,
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
				return "", fmt.Errorf("could not delete batch from world state. %s", err)
			}
		}
	}

	// Validate output batch
	isValidBatch, err := validateBatch(ctx, outputBatch.ID, outputBatch.ProductionUnitID, outputBatch.BatchInternalID, outputBatch.SupplierID, string(outputBatch.Unit), string(outputBatch.BatchType), outputBatch.BatchComposition, outputBatch.Quantity, outputBatch.ECS, outputBatch.SES, outputBatch.IsInTransit)
	if !isValidBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %w", err)
	}

	// Instantiate production
	production := &Production{
		DocType:           "p",
		ID:                productionID,
		ProductionUnitID:  outputBatch.ProductionUnitID,
		CompanyID:         companyID,
		ProductionType:    validProductionType,
		ActivityStartDate: civilDates[0],
		ActivityEndDate:   civilDates[1],
		ECS:               ECS,
		SES:               SES,
		InputBatches:      auxInputBatches,
		OutputBatch:       outputBatch,
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
func (c *StvgdContract) ReadProduction(ctx contractapi.TransactionContextInterface, productionID string) (*Production, error) {

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
	production := new(Production)
	// Unmarshal productionBytes to JSON
	err = json.Unmarshal(productionBytes, production)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Production")
	}

	return production, nil
}

// ! GetAllProductions returns all productions found in world state
func (c *StvgdContract) GetAllProductions(ctx contractapi.TransactionContextInterface) ([]*Production, error) {
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
