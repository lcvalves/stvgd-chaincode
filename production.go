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
	ObjectType        string             `json:"docType"` // docType ("p") is used to distinguish the various types of objects in state database
	ID                string             `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string             `json:"productionUnitID"`
	CompanyID         string             `json:"companyID"`
	ProductionTypeID  ProductionType     `json:"productionTypeID"`
	InputBatches      map[string]float32 `json:"inputBatches"`
	OutputBatch       Batch              `json:"outputBatch"`
	ActivityStartDate civil.DateTime     `json:"activityStartDate"`
	ActivityEndDate   civil.DateTime     `json:"activityEndDate"`
	ECS               float32            `json:"ecs"`
	SES               float32            `json:"ses"`
}

/*
 * -----------------------------------
 * TRANSACTIONS
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
func (c *StvgdContract) CreateProduction(ctx contractapi.TransactionContextInterface, productionID, productionUnitID, companyID, productionTypeID, activityStartDate, activityEndDate string,
	inputBatches map[string]float32, outputBatch Batch, ECS, SES float32) (string, error) {

	// Checks if the production activity ID already exists
	exists, err := c.ProductionExists(ctx, productionID)
	if err != nil {
		return "", fmt.Errorf("could not read production activity from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("production activity [%s] already exists", productionID)
	}

	// Checks equality in production IDs & production units
	if productionUnitID != outputBatch.ProductionUnitID {
		return "", fmt.Errorf("production unit's ID [%s] must be the same as output batch's production unit's ID [%s]", productionUnitID, outputBatch.ProductionUnitID)
	}

	// Validate production type
	validProductionType, err := validateProductionType(productionTypeID)
	if err != nil {
		return "", fmt.Errorf("could not validate activity type. %s", err)
	}

	// Validate dates
	civilDates, err := validateDates(activityStartDate, activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates. %s", err)
	}

	// Input batches min length (1)
	if len(inputBatches) < 1 {
		return "", fmt.Errorf("production must have atleast 1 input batch")
	}

	// Validate scores
	validScores, err := validateScores(ECS, SES)
	if !validScores {
		return "", fmt.Errorf("invalid scores. %s", err)
	}

	// Aux variables
	activities := make([]string, 0)
	parentBatches := make([]string, 0)

	for batchID, quantity := range inputBatches { // In every single input batch

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

		// Validate inserted quantities (0 <= quantity(inputBatch) <= batch.Quantity)
		switch {
		case quantity <= 0:
			return "", fmt.Errorf("input batches' quantities must be greater than 0 (input quantity for [%s] is %.2f)", batchID, quantity)
		case quantity > batch.Quantity:
			return "", fmt.Errorf("input batches' quantities must not exceed the batch's total quantity ([%s] max quantity is %.2f)", batchID, batch.Quantity)
		}

		// Append input batches traceability for output batch
		activities = append(activities, batch.Traceability.Activities...)
		parentBatches = append(parentBatches, batchID)

		// Initialize updated/"new" Batch object
		updatedInputBatch := &Batch{
			ObjectType:       batch.ObjectType,
			ID:               batch.ID,
			BatchTypeID:      batch.BatchTypeID,
			ProductionUnitID: batch.ProductionUnitID,
			BatchInternalID:  batch.BatchInternalID,
			SupplierID:       batch.SupplierID,
			BatchComposition: batch.BatchComposition,
			Traceability:     batch.Traceability,
			Quantity:         batch.Quantity - quantity,
			Unit:             batch.Unit,
			ECS:              batch.ECS,
			SES:              batch.SES,
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

	// Validate output batch
	validBatch, err := validateBatch(ctx, outputBatch.ID, outputBatch.ProductionUnitID, outputBatch.BatchInternalID, outputBatch.SupplierID,
		string(outputBatch.Unit), string(outputBatch.BatchTypeID), outputBatch.BatchComposition, outputBatch.Quantity, outputBatch.ECS, outputBatch.SES)
	if !validBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %v", err)
	}

	// Setup output batch Traceability
	activities = append(activities, productionID)
	outputBatch.Traceability = Traceability{
		Activities:    activities,
		ParentBatches: parentBatches,
	}

	// Instatiate production
	production := &Production{
		ObjectType:        "p",
		ID:                productionID,
		ProductionUnitID:  productionUnitID,
		CompanyID:         companyID,
		ProductionTypeID:  validProductionType,
		InputBatches:      inputBatches,
		OutputBatch:       outputBatch,
		ActivityStartDate: civilDates[0],
		ActivityEndDate:   civilDates[1],
		ECS:               ECS,
		SES:               SES,
	}

	// Marshal batch to bytes
	batchBytes, err := json.Marshal(outputBatch)
	if err != nil {
		return "", err
	}
	// Put batchBytes in world state
	err = ctx.GetStub().PutState(outputBatch.ID, batchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put batch to world state: %v", err)
	}

	// Marshal production to bytes
	productionBytes, err := json.Marshal(production)
	if err != nil {
		return "", err
	}
	// Put productionBytes in world state
	err = ctx.GetStub().PutState(productionID, productionBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put production to world state: %v", err)
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

//! GetAllProductions returns all productions found in world state
func (c *StvgdContract) GetAllProductions(ctx contractapi.TransactionContextInterface) ([]*Production, error) {
	// range query with empty string for endKey does an open-ended query of all productions in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("p", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var productions []*Production
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var production Production
		err = json.Unmarshal(queryResponse.Value, &production)
		if err != nil {
			return nil, err
		}
		productions = append(productions, &production)
	}

	return productions, nil
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

//! DeleteAllProductions deletes all production found in world state
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
