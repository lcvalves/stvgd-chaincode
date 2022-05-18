/*
 * SPDX-License-Identifier: Apache-2.0
 */

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

//TODO: Init LogisticalActivityTransport with return = false;

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
} */

/*
 * -----------------------------------
 * BATCH
 * -----------------------------------
 */

// BatchExists returns true when batch with given ID exists in world state
func (c *StvgdContract) BatchExists(ctx contractapi.TransactionContextInterface, batchID string) (bool, error) {
	data, err := ctx.GetStub().GetState(batchID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateBatch creates a new instance of Batch
func (c *StvgdContract) CreateBatch(ctx contractapi.TransactionContextInterface, batchID, batchTypeID, productionActivityID, productionUnitID, batchInternalID, supplierID, unit string, batchComposition map[string]float32, quantity, ecs, ses float32) (string, error) {

	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("the batch [%s] already exists", batchID)
	}

	if quantity < 0 {
		return "", fmt.Errorf("the batch quantity should be greater than 0")
	}

	batch := &Batch{
		ObjectType:           "batch",
		ID:                   batchID,
		BatchTypeID:          batchTypeID,
		ProductionActivityID: productionActivityID,
		ProductionUnitID:     productionUnitID,
		BatchInternalID:      batchInternalID,
		SupplierID:           supplierID,
		BatchComposition:     batchComposition,
		Quantity:             quantity,
		Unit:                 unit,
		ECS:                  ecs,
		SES:                  ses,
	}

	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(batch.ID, batchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to add batch to world state: %v", err)
	}

	return fmt.Sprintf("[%s] created successfully", batchID), nil
}

// ReadBatch retrieves an instance of Batch from the world state
func (c *StvgdContract) ReadBatch(ctx contractapi.TransactionContextInterface, batchID string) (*Batch, error) {

	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("[%s] does not exist", batchID)
	}

	batchBytes, _ := ctx.GetStub().GetState(batchID)

	batch := new(Batch)

	err = json.Unmarshal(batchBytes, batch)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Batch")
	}

	return batch, nil
}

// constructQueryResponseFromIterator constructs a slice of batches from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Batch, error) {
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

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Batch, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// GetAllBatches queries for all batches.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (docType).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (c *StvgdContract) GetAllBatches(ctx contractapi.TransactionContextInterface) ([]*Batch, error) {
	queryString := `{"selector":{"docType":"batch"}}`
	return getQueryResultForQueryString(ctx, queryString)
}

/*
// GetAssetHistory returns the chain of custody for a batch since issuance.
func (c *StvgdContract) GetBatchHistory(ctx contractapi.TransactionContextInterface, batchID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", batchID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(batchID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var batch Batch
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &batch)
			if err != nil {
				return nil, err
			}
		} else {
			batch = Batch{
				ID: batchID,
			}
		}

		timestamp := timestamppb.New(response.Timestamp.AsTime())
		if timestamp.CheckValid() != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp.AsTime(),
			Record:    &batch,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}
*/

// UpdateBatchQuantity updates the quantity of a Batch from the world state
func (c *StvgdContract) UpdateBatchQuantity(ctx contractapi.TransactionContextInterface, batchID string, newQuantity float32) (string, error) {

	// Verifies if Batch that has batchID already exists
	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("%s does not exist", batchID)
	}

	outdatedBatchBytes, _ := ctx.GetStub().GetState(batchID) // Gets "old" Batch bytes from batchID

	outdatedBatch := new(Batch) // Initialize outdated/"old" Batch object

	// Parses the JSON-encoded data in bytes (outdatedBatchBytes) to the "old" Batch object (outdatedBatch)
	err = json.Unmarshal(outdatedBatchBytes, outdatedBatch)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal batch world state data to type Batch")
	}

	// Checks if quantity >= 0
	if newQuantity < 0 {
		return "", fmt.Errorf("the new quantity should be greater than 0")
	} else {
		// Initialize updated/"new" Batch object
		updatedBatch := &Batch{
			ObjectType:           outdatedBatch.ObjectType,
			ID:                   outdatedBatch.ID,
			BatchTypeID:          outdatedBatch.BatchTypeID,
			ProductionActivityID: outdatedBatch.ProductionActivityID,
			ProductionUnitID:     outdatedBatch.ProductionUnitID,
			BatchInternalID:      outdatedBatch.BatchInternalID,
			SupplierID:           outdatedBatch.SupplierID,
			BatchComposition:     outdatedBatch.BatchComposition,
			Quantity:             newQuantity,
			Unit:                 outdatedBatch.Unit,
			ECS:                  outdatedBatch.ECS,
			SES:                  outdatedBatch.SES,
		}

		updatedBatchBytes, _ := json.Marshal(updatedBatch) // Encodes the JSON updatedBatch data to bytes

		err = ctx.GetStub().PutState(batchID, updatedBatchBytes) // Updates world state with newly updated Batch
		if err != nil {
			return "", fmt.Errorf("could not write batch to world state. %s", err)
		} else if newQuantity == 0 { // Deletes the batch if there is no more quantity left / newQuantity = 0
			_, err = c.DeleteBatch(ctx, batchID)
			if err != nil {
				return "", fmt.Errorf("could not delete batch from world state. %s", err)
			} else {
				return fmt.Sprintf("[%s]'s quantity was successfully updated to %.2f%s and deleted from world state", batchID, newQuantity, outdatedBatch.Unit), nil
			}
		} else {
			return fmt.Sprintf("[%s]'s quantity was successfully updated to %.2f%s", batchID, newQuantity, outdatedBatch.Unit), nil
		}
	}
}

// TransferBatch transfers a batch by setting a new production unit id on the batch
func (c *StvgdContract) TransferBatch(ctx contractapi.TransactionContextInterface, batchID, newProductionUnitID string) (string, error) {

	// Verifies if Batch that has batchID already exists
	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("[%s] does not exist", batchID)
	}

	outdatedBatchBytes, _ := ctx.GetStub().GetState(batchID) // Gets "old" Batch bytes from batchID

	outdatedBatch := new(Batch) // Initialize outdated/"old" Batch object

	// Parses the JSON-encoded data in bytes (outdatedBatchBytes) to the "old" Batch object (outdatedBatch)
	err = json.Unmarshal(outdatedBatchBytes, outdatedBatch)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal batch world state data to type Batch")
	}

	// Checks if new owner is different
	if newProductionUnitID == outdatedBatch.ProductionUnitID {
		return "", fmt.Errorf("cannot transfer a batch to the current owner/production unit %s[]", outdatedBatch.ProductionUnitID)
	} else {
		// Initialize updated/"new" Batch object
		updatedBatch := &Batch{
			ObjectType:           outdatedBatch.ObjectType,
			ID:                   outdatedBatch.ID,
			BatchTypeID:          outdatedBatch.BatchTypeID,
			ProductionActivityID: outdatedBatch.ProductionActivityID,
			ProductionUnitID:     newProductionUnitID,
			BatchInternalID:      outdatedBatch.BatchInternalID,
			SupplierID:           outdatedBatch.SupplierID,
			BatchComposition:     outdatedBatch.BatchComposition,
			Quantity:             outdatedBatch.Quantity,
			Unit:                 outdatedBatch.Unit,
			ECS:                  outdatedBatch.ECS,
			SES:                  outdatedBatch.SES,
		}

		updatedBatchBytes, _ := json.Marshal(updatedBatch) // Encodes the JSON updatedBatch data to bytes

		err = ctx.GetStub().PutState(batchID, updatedBatchBytes) // Updates world state with newly updated Batch
		if err != nil {
			return "", fmt.Errorf("could not write to world state. %s", err)
		} else {
			return fmt.Sprintf("[%s] transfered successfully to production unit [%s]", batchID, newProductionUnitID), nil
		}
	}
}

// DeleteBatch deletes an instance of Batch from the world state
func (c *StvgdContract) DeleteBatch(ctx contractapi.TransactionContextInterface, batchID string) (string, error) {
	exists, err := c.BatchExists(ctx, batchID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("[%s] does not exist", batchID)
	}

	err = ctx.GetStub().DelState(batchID)
	if err != nil {
		return "", fmt.Errorf("could not delete batch from world state. %s", err)
	} else {
		return fmt.Sprintf("[%s] deleted successfully", batchID), nil
	}
}

// DeleteAllBatches deletes all batches found in world state
func (c *StvgdContract) DeleteAllBatches(ctx contractapi.TransactionContextInterface) (string, error) {

	batches, err := c.GetAllBatches(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if len(batches) == 0 {
		return "", fmt.Errorf("there are no batches in world state to delete")
	}

	for _, batch := range batches {
		err = ctx.GetStub().DelState(batch.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete batch from world state. %s", err)
		}
	}

	return "all the batches were successfully deleted", nil
}

/*
 * -----------------------------------
 * PRODUCTION ACTIVITY
 * -----------------------------------
 */

// ProductionActivityExists returns true when productionActivity with given ID exists in world state
func (c *StvgdContract) ProductionActivityExists(ctx contractapi.TransactionContextInterface, productionActivityID string) (bool, error) {
	data, err := ctx.GetStub().GetState(productionActivityID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateProductionActivity creates a new instance of ProductionActivity
func (c *StvgdContract) CreateProductionActivity(ctx contractapi.TransactionContextInterface, productionActivityID, productionUnitID, companyID, activityTypeID, activityStartDate, activityEndDate string, inputBatches map[string]float32, outputBatch Batch, ECS, SES float32) (string, error) {

	// Checks if the output lot ID already exists
	exists, err := c.BatchExists(ctx, outputBatch.ID)
	if err != nil {
		return "", fmt.Errorf("could not read batch from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("[%s] already exists", outputBatch.ID)
	}

	// Checks if the production activity ID already exists
	exists, err = c.ProductionActivityExists(ctx, productionActivityID)
	if err != nil {
		return "", fmt.Errorf("could not read production activity from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("production activity [%s] already exists", productionActivityID)
	}

	// Checks equality in production activity IDs & production units
	if productionActivityID != outputBatch.ProductionActivityID {
		return "", fmt.Errorf("production activity's ID [%s] must be the same as output batch's production activity's ID [%s]", productionActivityID, outputBatch.ProductionActivityID)
	} else if productionUnitID != outputBatch.ProductionUnitID {
		return "", fmt.Errorf("production unit's ID [%s] must be the same as output batch's production unit's ID [%s]", productionUnitID, outputBatch.ProductionUnitID)
	}

	// Date parsing
	civilActivityStartDate, err := civil.ParseDateTime(activityStartDate)
	if err != nil {
		return "", fmt.Errorf("could not parse the activity start date to correct format. %s", err)
	}
	civilActivityEndDate, err := civil.ParseDateTime(activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not parse the activity end date to correct format. %s", err)
	}

	// Checks if the sent date is before received date
	if civilActivityStartDate.After(civilActivityEndDate) {
		return "", fmt.Errorf("activity start date can't be after the activity end date")
	}

	// Input batches audit
	if len(inputBatches) > 0 { // If production activity uses input batches

		for batchID, quantity := range inputBatches { // In every single input batch

			// Checks if the batch ID exists in world state
			exists, err := c.BatchExists(ctx, batchID)
			if err != nil {
				return "", fmt.Errorf("could not read batch from world state. %s", err)
			} else if !exists {
				return "", fmt.Errorf("[%s] does not exist", batchID)
			}

			// Reads the batch
			batch, err := c.ReadBatch(ctx, batchID)
			if err != nil {
				return "", fmt.Errorf("could not read batch from world state. %s", err)
			}

			// Validate inserted quantities (0 <= quantity(inputBatch) <= batch.Quantity)
			switch {
			case quantity <= 0:
				return "", fmt.Errorf("input batches' quantities must be greater than 0 (input quantity for batch [%s] is %.2f)", batchID, quantity)
			case quantity > batch.Quantity:
				return "", fmt.Errorf("input batches' quantities must not exceed the batch's total quantity (batch [%s] max quantity is %.2f)", batchID, batch.Quantity)
			}

			// Subtract batch's quantity with input batches' quantity //! CURRENTLY NOT WORKING
			_, err = c.UpdateBatchQuantity(ctx, batchID, batch.Quantity-quantity)
			if err != nil {
				return "", fmt.Errorf("could not write batch to world state. %s", err)
			}

			// Transfer input batches ownership to new production unit / owner
			if batch.ProductionUnitID != productionUnitID { // Only transfer if production units for the input batches are different
				_, err = c.TransferBatch(ctx, batchID, productionUnitID)
				if err != nil {
					return "", fmt.Errorf("could not write batch to world state. %s", err)
				}
			}
		}

	}

	// Create production activity's output batch
	_, err = c.CreateBatch(ctx, outputBatch.ID, outputBatch.BatchTypeID, productionActivityID, outputBatch.ProductionUnitID, outputBatch.BatchInternalID, outputBatch.SupplierID, outputBatch.Unit, outputBatch.BatchComposition, outputBatch.Quantity, outputBatch.ECS, outputBatch.SES)
	if err != nil {
		return "", fmt.Errorf("could not write batch to world state. %s", err)
	}

	productionActivity := &ProductionActivity{
		ObjectType:        "pa",
		ID:                productionActivityID,
		ProductionUnitID:  productionUnitID,
		CompanyID:         companyID,
		ActivityTypeID:    activityTypeID,
		InputBatches:      inputBatches,
		OutputBatch:       outputBatch,
		ActivityStartDate: civilActivityStartDate,
		ActivityEndDate:   civilActivityEndDate,
		ECS:               ECS,
		SES:               SES,
	}

	productionActivityBytes, err := json.Marshal(productionActivity)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(productionActivityID, productionActivityBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put production activity to world state: %v", err)
	}

	return fmt.Sprintf("production activity [%s] & [%s] were successfully added to the ledger", productionActivityID, outputBatch.ID), nil
}

// ReadProductionActivity retrieves an instance of ProductionActivity from the world state
func (c *StvgdContract) ReadProductionActivity(ctx contractapi.TransactionContextInterface, productionActivityID string) (*ProductionActivity, error) {

	exists, err := c.ProductionActivityExists(ctx, productionActivityID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the production activity [%s] does not exist", productionActivityID)
	}

	productionActivityBytes, _ := ctx.GetStub().GetState(productionActivityID)

	productionActivity := new(ProductionActivity)

	err = json.Unmarshal(productionActivityBytes, productionActivity)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type ProductionActivity")
	}

	return productionActivity, nil
}

// GetAllProdActivities returns all production activities found in world state
func (c *StvgdContract) GetAllProdActivities(ctx contractapi.TransactionContextInterface) ([]*ProductionActivity, error) {
	// range query with empty string for endKey does an
	// open-ended query of all production activities in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("pa", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var prodActivities []*ProductionActivity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var productionActivity ProductionActivity
		err = json.Unmarshal(queryResponse.Value, &productionActivity)
		if err != nil {
			return nil, err
		}
		prodActivities = append(prodActivities, &productionActivity)
	}

	return prodActivities, nil
}

// DeleteAllProdActivities deletes all production activities found in world state
func (c *StvgdContract) DeleteAllProdActivities(ctx contractapi.TransactionContextInterface) (string, error) {

	prodActivities, err := c.GetAllProdActivities(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if len(prodActivities) == 0 {
		return "", fmt.Errorf("there are no productions activities in world state to delete")
	}

	for _, productionActivity := range prodActivities {
		err = ctx.GetStub().DelState(productionActivity.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete from world state. %s", err)
		}
	}

	return "all the production activities were successfully deleted", nil
}

/*
 * -----------------------------------
 * LOGISTICS ACTIVITY
 * -----------------------------------
 */
/*
// LogActivityExists returns true when logActivity with given ID exists in world state
func (c *StvgdContract) LogActivityExists(ctx contractapi.TransactionContextInterface, logActivityID string) (bool, error) {
	data, err := ctx.GetStub().GetState(logActivityID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateProdActivity creates a new instance of ProdActivity
func (c *StvgdContract) CreateLogActivity(ctx contractapi.TransactionContextInterface, logActivityID, transportationType,
	prodUnitFrom, prodUnitTo string, lots []string, distance, cost float32, dateSent, dateReceived string, envScore float32) (string, error) {

	// Checks if the logistic activity ID already exists
	exists, err := c.LogActivityExists(ctx, logActivityID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("the logistic activity [%s] already exists", logActivityID)
	}

	// Checks if the origin & destination production units are different
	if prodUnitFrom == prodUnitTo {
		return "", fmt.Errorf("origin & destination production units can't be the the same")
	}

	// Checks if the distance is not 0
	if distance <= 0 {
		return "", fmt.Errorf("distance can't be 0")
	}

	// Checks if the cost is not 0
	if cost <= 0 {
		return "", fmt.Errorf("cost can't be 0")
	}

	// Date parsing
	civilDateSent, err := civil.ParseDate(dateSent)
	if err != nil {
		return "", fmt.Errorf("could not parse the sent date to correct format. %s", err)
	}
	civilDateReceived, err := civil.ParseDate(dateReceived)
	if err != nil {
		return "", fmt.Errorf("could not parse the received date to correct format. %s", err)
	}

	// Checks if the sent date is before received date
	if civilDateSent.After(civilDateReceived) {
		return "", fmt.Errorf("sent date can't be after the received date")
	}

	// Lots audit
	if len(lots) <= 0 { // force atleast 1 lot per logistic activity
		return "", fmt.Errorf("the logistic activity must trasnport atleast 1 lot")
	} else {

		for _, lotID := range lots { // For every lot

			// Check if the lot ID exists in world state
			exists, err := c.LotExists(ctx, lotID)
			if err != nil {
				return "", fmt.Errorf("could not read from world state. %s", err)
			} else if !exists {
				return "", fmt.Errorf("the lot [%s] does not exist", lotID)
			}

			// Reads the lot
			lot, err := c.ReadLot(ctx, lotID)
			if err != nil {
				return "", fmt.Errorf("could not read from world state. %s", err)
			}

			// Checks equality in logistic activity's origin production unit & lot's production unit
			if prodUnitFrom != lot.ProdUnit {
				return "", fmt.Errorf("logistic activity's origin production unit [%s] must be the same as the lot's [%s] production unit [%s]", prodUnitFrom, lotID, lot.ProdUnit)
			} else { // Transfer lots ownership to new production unit / owner
				_, err = c.TransferLot(ctx, lotID, prodUnitTo)
				if err != nil {
					return "", fmt.Errorf("could not write to world state. %s", err)
				}
			}

		}
	}

	logActivity := &LogActivity{
		DocType:            "logActivity",
		ID:                 logActivityID,
		TransportationType: transportationType,
		ProdUnitFrom:       prodUnitFrom,
		ProdUnitTo:         prodUnitTo,
		Lots:               lots,
		Distance:           distance,
		Cost:               cost,
		DateSent:           dateSent,
		DateReceived:       dateReceived,
		EnvScore:           envScore,
	}

	logActivityBytes, err := json.Marshal(logActivity)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(logActivityID, logActivityBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put to world state: %v", err)
	}

	return fmt.Sprintf("logistic activity [%s] was successfully added to the ledger", logActivityID), nil
}

// ReadProdActivity retrieves an instance of ProdActivity from the world state
func (c *StvgdContract) ReadLogActivity(ctx contractapi.TransactionContextInterface, logActivityID string) (*LogActivity, error) {

	exists, err := c.ProdActivityExists(ctx, logActivityID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the logistic activity [%s] does not exist", logActivityID)
	}

	logActivityBytes, _ := ctx.GetStub().GetState(logActivityID)

	logActivity := new(LogActivity)

	err = json.Unmarshal(logActivityBytes, logActivity)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type LogActivity")
	}

	return logActivity, nil
}

// constructQueryResponseFromIterator constructs a slice of lots from the resultsIterator
func constructLogActivityQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*LogActivity, error) {
	var logActivities []*LogActivity
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var logActivity LogActivity
		err = json.Unmarshal(queryResult.Value, &logActivity)
		if err != nil {
			return nil, err
		}
		logActivities = append(logActivities, &logActivity)
	}

	return logActivities, nil
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getLogActivityQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*LogActivity, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructLogActivityQueryResponseFromIterator(resultsIterator)
}

// GetAllLogActivities queries for all logistic activities.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (docType).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (c *StvgdContract) GetAllLogActivities(ctx contractapi.TransactionContextInterface) ([]*LogActivity, error) {
	queryString := `{"selector":{"docType":"logActivity"}}`
	return getLogActivityQueryResultForQueryString(ctx, queryString)
}

// DeleteAllLogActivities deletes all production activities found in world state
func (c *StvgdContract) DeleteAllLogActivities(ctx contractapi.TransactionContextInterface) (string, error) {

	logActivities, err := c.GetAllLogActivities(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if len(logActivities) == 0 {
		return "", fmt.Errorf("there are no logistic activities in world state to delete")
	}

	for _, logActivity := range logActivities {
		err = ctx.GetStub().DelState(logActivity.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete from world state. %s", err)
		}
	}

	return "all the logistic activities were successfully deleted", nil
}
*/

//TODO: ReceiveBatch
//Updates id & internal-id
