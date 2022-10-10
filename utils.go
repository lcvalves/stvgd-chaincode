package main

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// StvgdContract contract for managing CRUD for Stvgd
type StvgdContract struct {
	contractapi.Contract
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Batch    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

/*
 * -----------------------------------
 - AUX Functions
 * -----------------------------------
*/

// Date validation
func validateDates(startDate, endDate string) ([]civil.DateTime, error) {
	// Date parsing
	civilActivityStartDate, err := civil.ParseDateTime(startDate)
	if err != nil {
		return nil, fmt.Errorf("could not parse the activity start date to correct format. %s", err)
	}
	civilActivityEndDate, err := civil.ParseDateTime(endDate)
	if err != nil {
		return nil, fmt.Errorf("could not parse the activity end date to correct format. %s", err)
	}

	// Checks if the sent date is before received date
	if civilActivityStartDate.After(civilActivityEndDate) {
		return nil, fmt.Errorf("activity start date can't be after the activity end date")
	}

	return []civil.DateTime{civilActivityStartDate, civilActivityEndDate}, nil
}

// Scores validation
func validateScores(ecs, ses float32) (bool, error) {
	// Range validate ECS
	if ecs < -10.0 || ecs > 10.0 {
		return false, fmt.Errorf("environmental & circular score out of bounds (should be between -10 & 10)")
	}
	// Range validate SES
	if ses < -10.0 || ses > 10.0 {
		return false, fmt.Errorf("social & economic score out of bounds (should be between -10 & 10)")
	}
	return true, nil
}

/*
 * -----------------------------------
 - ACTIVITIES Validation
 * -----------------------------------
*/

// Traceability Activity type validation
func validateActivityType(activityID string) (string, error) {
	var activityPrefix string
	switch activityID[0:1] {
	case "r":
		switch activityID[1:3] {
		case "g-":
			activityPrefix = "rg"
		case "c-":
			activityPrefix = "rc"
		}
	case "p":
		switch activityID[1:2] {
		case "-":
			activityPrefix = "p"
		}
	case "t":
		switch activityID[1:2] {
		case "-":
			activityPrefix = "t"
		}
	default:
		return "", fmt.Errorf("incorrect activity prefix")
	}

	return activityPrefix, nil
}

// Production type validation
func validateProductionType(productionTypeID string) (ProductionType, error) {
	var productionType ProductionType
	switch productionTypeID {
	case "SPINNING":
		productionType = Spinning
	case "WEAVING":
		productionType = Weaving
	case "KNITTING":
		productionType = Knitting
	case "DYEING_FINISHING":
		productionType = DyeingFinishing
	case "CONFECTION":
		productionType = Confection
	default:
		return "", fmt.Errorf("production type not found")
	}

	return productionType, nil
}

// Transportation type validation
func validateTransportationType(transportationTypeID string) (TransportationType, error) {
	var transportationType TransportationType
	switch transportationTypeID {
	case "ROAD":
		transportationType = Road
	case "MARITIME":
		transportationType = Maritime
	case "AIR":
		transportationType = Air
	case "RAIL":
		transportationType = Rail
	case "INTERMODAL":
		transportationType = Intermodal
	default:
		return "", fmt.Errorf("transportation type not found")
	}

	return transportationType, nil
}

/*
 * -----------------------------------
 - BATCH Validation
 * -----------------------------------
*/

// validateBatch validates batch for correct inputs/fields on Registration & Production activities
func validateBatch(ctx contractapi.TransactionContextInterface, batchID, productionUnitID, batchInternalID, supplierID, unit, batchType string, batchComposition map[string]float32, quantity, ecs, ses float32, isInTransit bool) (bool, error) {

	// Batch prefix validation
	switch batchID[0:2] {
	case "b-":
	default:
		return false, fmt.Errorf("incorrect batch prefix. (should be [b-...])")
	}

	// Verifies if Batch has a batchID that already exists
	data, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return false, fmt.Errorf("could not read batch from world state: %w", err)
	} else if data != nil {
		return false, fmt.Errorf("batch [%s] already exists", batchID)
	}

	// Validate batch internal ID
	if batchInternalID == "" {
		return false, fmt.Errorf("batch internal ID must not be empty")
	}

	// Validate supplier ID
	if supplierID == "" {
		return false, fmt.Errorf("supplier ID must not be empty")
	}

	// Validate supplier ID
	if isInTransit {
		return false, fmt.Errorf("batch must not be in transit")
	}

	// Validate batch type
	switch batchType {
	case "FIBER":
	case "YARN":
	case "MESH":
	case "FABRIC":
	case "DYED_MESH":
	case "FINISHED_MESH":
	case "DYED_FABRIC":
	case "FINISHED_FABRIC":
	case "CUT":
	case "FINISHED_PIECE":
	case "OTHER":
	default:
		return false, fmt.Errorf("batch type is not valid")
	}

	// Validate batch composition
	var percentageSum float32 = 0.00 // Local variable for percentage sum validation
	for _, percentage := range batchComposition {
		percentageSum += percentage
		if percentageSum > 100 {
			return false, fmt.Errorf(" batch composition percentage sum should be equal to 100")
		}
	}
	if percentageSum != 100 {
		return false, fmt.Errorf("batch composition percentage sum should be equal to 100")
	}

	// Validate quantity
	if quantity < 0 {
		return false, fmt.Errorf("batch quantity should be 0+")
	}

	// Validate unit
	switch unit {
	case "KG":
	case "L":
	case "M":
	case "M2":
	default:
		return false, fmt.Errorf("unit is not valid")
	}

	// Validate scores (-10 <= ECS & SES <= 10)
	validScores, err := validateScores(ecs, ses)
	if !validScores {
		return false, fmt.Errorf("invalid scores: %w", err)
	}

	return true, nil
}

/*
 * -----------------------------------
 - RICH QUERIES - Batch
 * -----------------------------------
*/

// getQueryResultForQueryString_batch executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringBatch(ctx contractapi.TransactionContextInterface, queryString string) ([]*Batch, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorBatch(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of batches from the resultsIterator
func constructQueryResponseFromIteratorBatch(resultsIterator shim.StateQueryIteratorInterface) ([]*Batch, error) {
	var batches []*Batch
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var batch Batch
		err = json.Unmarshal(queryResult.Value, &batch)
		if err != nil {
			return nil, err
		}
		batches = append(batches, &batch)
	}

	return batches, nil
}

/*
 * -----------------------------------
 - RICH QUERIES - Registration
 * -----------------------------------
*/

// getQueryResultForQueryStringRegistration executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringRegistration(ctx contractapi.TransactionContextInterface, queryString string) ([]*Registration, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorRegistration(resultsIterator)
}

// constructQueryResponseFromIteratorRegistration constructs a slice of registrations from the resultsIterator
func constructQueryResponseFromIteratorRegistration(resultsIterator shim.StateQueryIteratorInterface) ([]*Registration, error) {
	var registrations []*Registration
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var registration Registration
		err = json.Unmarshal(queryResult.Value, &registration)
		if err != nil {
			return nil, err
		}
		registrations = append(registrations, &registration)
	}

	return registrations, nil
}

/*
 * -----------------------------------
 - RICH QUERIES - Production
 * -----------------------------------
*/

// getQueryResultForQueryStringProduction executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringProduction(ctx contractapi.TransactionContextInterface, queryString string) ([]*Production, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorProduction(resultsIterator)
}

// constructQueryResponseFromIteratorProduction constructs a slice of production activities from the resultsIterator
func constructQueryResponseFromIteratorProduction(resultsIterator shim.StateQueryIteratorInterface) ([]*Production, error) {
	var productions []*Production
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var production Production
		err = json.Unmarshal(queryResult.Value, &production)
		if err != nil {
			return nil, err
		}
		productions = append(productions, &production)
	}

	return productions, nil
}

/*
 * -----------------------------------
 - RICH QUERIES - Transport
 * -----------------------------------
*/

// getQueryResultForQueryStringTransport executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringTransport(ctx contractapi.TransactionContextInterface, queryString string) ([]*Transport, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorTransport(resultsIterator)
}

// constructQueryResponseFromIteratorTransport constructs a slice of batches from the resultsIterator
func constructQueryResponseFromIteratorTransport(resultsIterator shim.StateQueryIteratorInterface) ([]*Transport, error) {
	var transports []*Transport
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transport Transport
		err = json.Unmarshal(queryResult.Value, &transport)
		if err != nil {
			return nil, err
		}
		transports = append(transports, &transport)
	}

	return transports, nil
}

/*
 * -----------------------------------
 - RICH QUERIES - Reception
 * -----------------------------------
*/

// getQueryResultForQueryStringReception executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringReception(ctx contractapi.TransactionContextInterface, queryString string) ([]*Reception, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorReception(resultsIterator)
}

// constructQueryResponseFromIteratorReception constructs a slice of receptions from the resultsIterator
func constructQueryResponseFromIteratorReception(resultsIterator shim.StateQueryIteratorInterface) ([]*Reception, error) {
	var receptions []*Reception
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var reception Reception
		err = json.Unmarshal(queryResult.Value, &reception)
		if err != nil {
			return nil, err
		}
		receptions = append(receptions, &reception)
	}

	return receptions, nil
}
