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
// InitLedger adds a base set of ProdActivities to the ledger
func (c *StvgdContract) InitLedger(ctx contractapi.TransactionContextInterface) (string, error) {
	//TODO: Simulate supply chain use case with proper instances of each struct
	// (check with other STVGD participants)
	lots := []Lot{
		{ObjectType: "lot", ID: "lot01", LotType: "test-type", ProdActivity: "pa01", Amount: 100, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot01-iid01"},
		{ObjectType: "lot", ID: "lot02", LotType: "test-type", ProdActivity: "pa02", Amount: 200, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot02-iid01"},
		{ObjectType: "lot", ID: "lot03", LotType: "test-type", ProdActivity: "pa03", Amount: 300, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot03-iid01"},
		{ObjectType: "lot", ID: "lot04", LotType: "test-type", ProdActivity: "pa04", Amount: 400, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot04-iid01"},
		{ObjectType: "lot", ID: "lot05", LotType: "test-type", ProdActivity: "pa05", Amount: 500, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot05-iid01"},
		{ObjectType: "lot", ID: "lot06", LotType: "test-type", Amount: 600, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot06-iid01"},
		{ObjectType: "lot", ID: "lot07", LotType: "test-type", Amount: 700, Unit: "KG", ProdUnit: "punit03", LotInternalID: "lot07-iid01"},
	}
	prodActivities := []ProdActivity{
		{ObjectType: "prodActivity", ID: "pa01", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[0], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 73},
		{ObjectType: "prodActivity", ID: "pa02", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[1], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 16},
		{ObjectType: "prodActivity", ID: "pa03", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 20, "lot02": 15}, OutputLot: lots[2], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 51},
		{ObjectType: "prodActivity", ID: "pa04", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[3], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 26},
		{ObjectType: "prodActivity", ID: "pa05", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[4], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 14},
		{ObjectType: "prodActivity", ID: "pa06", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot04": 50, "lot05": 20}, OutputLot: lots[5], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 20},
		{ObjectType: "prodActivity", ID: "pa07", ActivityType: "test-type", ProdUnit: "punit03", InputLots: map[string]float32{"lot01": 30, "lot04": 10, "lot06": 10}, OutputLot: lots[6], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 100},
	}
	logActivities := []LogActivity{
		{ObjectType: "logActivity", ID: "la01", TransportationType: "test-type", ProdUnitFrom: "punit01", ProdUnitTo: "punit01", Lots: []string{"lot01"}, Distance: 10, Cost: 10, DateSent: "2022-01-01", DateReceived: "2022-02-01", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la02", TransportationType: "test-type", ProdUnitFrom: "punit01", ProdUnitTo: "punit01", Lots: []string{"lot01"}, Distance: 20, Cost: 20, DateSent: "2022-01-02", DateReceived: "2022-02-02", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la03", TransportationType: "test-type", ProdUnitFrom: "punit01", ProdUnitTo: "punit01", Lots: []string{"lot01"}, Distance: 30, Cost: 30, DateSent: "2022-01-03", DateReceived: "2022-02-03", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la04", TransportationType: "test-type", ProdUnitFrom: "punit02", ProdUnitTo: "punit02", Lots: []string{"lot01"}, Distance: 40, Cost: 40, DateSent: "2022-01-04", DateReceived: "2022-02-04", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la05", TransportationType: "test-type", ProdUnitFrom: "punit02", ProdUnitTo: "punit02", Lots: []string{"lot01"}, Distance: 50, Cost: 50, DateSent: "2022-01-05", DateReceived: "2022-02-05", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la06", TransportationType: "test-type", ProdUnitFrom: "punit02", ProdUnitTo: "punit02", Lots: []string{"lot04"}, Distance: 60, Cost: 60, DateSent: "2022-01-06", DateReceived: "2022-02-06", EnvScore: 50},
		{ObjectType: "logActivity", ID: "la07", TransportationType: "test-type", ProdUnitFrom: "punit03", ProdUnitTo: "punit03", Lots: []string{"lot01"}, Distance: 70, Cost: 70, DateSent: "2022-01-07", DateReceived: "2022-02-07", EnvScore: 50},
	}
	for _, lot := range lots {
		exists, err := c.LotExists(ctx, lot.ID)
		if err != nil {
			return "", fmt.Errorf("could not read from world state. %s", err)
		} else if exists {
			return "", fmt.Errorf("the lot %s already exists", lot.ID)
		}
		lotBytes, err := json.Marshal(lot)
		if err != nil {
			return "", err
		}
		err = ctx.GetStub().PutState(lot.ID, lotBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put to world state: %v", err)
		}
	}
	for _, prodActivity := range prodActivities {
		exists, err := c.ProdActivityExists(ctx, prodActivity.ID)
		if err != nil {
			return "", fmt.Errorf("could not read from world state. %s", err)
		} else if exists {
			return "", fmt.Errorf("the production activity [%s] already exists", prodActivity.ID)
		}
		prodActivityBytes, err := json.Marshal(prodActivity)
		if err != nil {
			return "", err
		}
		err = ctx.GetStub().PutState(prodActivity.ID, prodActivityBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put to world state: %v", err)
		}
	}
	for _, logActivity := range logActivities {
		exists, err := c.LogActivityExists(ctx, logActivity.ID)
		if err != nil {
			return "", fmt.Errorf("could not read from world state. %s", err)
		} else if exists {
			return "", fmt.Errorf("the logistic activity [%s] already exists", logActivity.ID)
		}
		logActivityBytes, err := json.Marshal(logActivity)
		if err != nil {
			return "", err
		}
		err = ctx.GetStub().PutState(logActivity.ID, logActivityBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put to world state: %v", err)
		}
	}
	return fmt.Sprintf("production activities [%s-%s], lots [%s-%s] and logistic activities [%s-%s] were successfully added to the ledger", prodActivities[0].ID, prodActivities[len(prodActivities)-1].ID, lots[0].ID, lots[len(lots)-1].ID, logActivities[0].ID, logActivities[len(logActivities)-1].ID), nil
}
*/

/*
 * AUX FUNCTIONS
 * ####################################
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

// Unit validation
func validateUnit(unitID string) (Unit, error) {
	var unit Unit
	switch unitID {
	case "KG":
		unit = Kilograms
	case "L":
		unit = Liters
	case "M":
		unit = Meters
	case "M2":
		unit = SquaredMeters
	default:
		return "", fmt.Errorf("unit not found")
	}

	return unit, nil
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

// Batch type validation
func validateBatchType(batchTypeID string) (BatchType, error) {
	var batchType BatchType
	switch batchTypeID {
	case "FIBER":
		batchType = Fiber
	case "YARN":
		batchType = Yarn
	case "MESH":
		batchType = Mesh
	case "FABRIC":
		batchType = Fabric
	case "DYED_MESH":
		batchType = DyedMesh
	case "FINISHED_MESH":
		batchType = FinishedMesh
	case "DYED_FABRIC":
		batchType = DyedFabric
	case "FINISHED_FABRIC":
		batchType = FinishedFabric
	case "CUT":
		batchType = Cut
	case "FINISHED_PIECE":
		batchType = FinishedPiece
	default:
		return "", fmt.Errorf("batch type not found")
	}

	return batchType, nil
}

/*
 * -----------------------------------
 - BATCH Validation
 * -----------------------------------
*/

// validateBatch validates batch for correct inputs/fields
func validateBatch(ctx contractapi.TransactionContextInterface, batchID, productionUnitID, batchInternalID, supplierID, unit, batchTypeID string,
	batchComposition map[string]float32, quantity, ecs, ses float32) (bool, error) {

	// Verifies if Batch has a batchID that already exists
	data, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return false, fmt.Errorf("could not read batch from world state. %s", err)
	} else if data != nil {
		return false, fmt.Errorf("batch [%s] already exists", batchID)
	}

	// Validate batch type
	validBatchType, err := validateBatchType(batchTypeID)
	if err != nil {
		return false, fmt.Errorf("could not validate batch type. %s", err)
	}

	// Validate batch composition
	var percentageSum float32 = 0.00 // Local variable for percentage sum validation
	for _, percentage := range batchComposition {
		percentageSum += percentage
		if percentageSum > 100 {
			return false, fmt.Errorf("the batch composition percentage sum should be equal to 100")
		}
	}
	if percentageSum != 100 {
		return false, fmt.Errorf("the batch composition percentage sum should be equal to 100")
	}

	// Validate quantity
	if quantity < 0 {
		return false, fmt.Errorf("batch quantity should be 0+")
	}

	// Validate unit
	validUnit, err := validateUnit(unit)
	if err != nil {
		return false, fmt.Errorf("could not validate unit. %s", err)
	}

	// Validate scores ( -10 <= ECS & SES <= 10)
	switch {
	case ecs <= -10 || ecs >= 10:
		return false, fmt.Errorf("ecs should be between -10 & 10")
	case ses <= -10 || ses >= 10:
		return false, fmt.Errorf("ecs should be between -10 & 10")
	}

	batch := &Batch{
		ObjectType:       "b",
		ID:               batchID,
		BatchTypeID:      validBatchType,
		ProductionUnitID: productionUnitID,
		BatchInternalID:  batchInternalID,
		SupplierID:       supplierID,
		BatchComposition: batchComposition,
		Quantity:         quantity,
		Unit:             validUnit,
		ECS:              ecs,
		SES:              ses,
	}

	return true, fmt.Errorf("batch [%s] validated", batch.ID)
}

/*
 * -----------------------------------
 - RICH QUERIES - Batch
 * -----------------------------------
*/
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

/*
 * -----------------------------------
 - RICH QUERIES - Production Activity
 * -----------------------------------
*/

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

/*
 * -----------------------------------
 - RICH QUERIES - Logistical Activities Transport
 * -----------------------------------
*/

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

/*
 * -----------------------------------
 - RICH QUERIES - Logistical Activities Registration
 * -----------------------------------
*/

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

/*
 * -----------------------------------
 - RICH QUERIES - Logistical Activities Reception
 * -----------------------------------
*/

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
