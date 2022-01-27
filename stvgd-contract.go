/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// StvgdContract contract for managing CRUD for Stvgd
type StvgdContract struct {
	contractapi.Contract
}

// InitLedger adds a base set of ProdActivities to the ledger
func (c *StvgdContract) InitLedger(ctx contractapi.TransactionContextInterface) (string, error) {

	lots := []Lot{
		{DocType: "lot", ID: "lot01", LotType: "test-type", ProdActivity: "pa01", Amount: 100, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot01-iid01"},
		{DocType: "lot", ID: "lot02", LotType: "test-type", ProdActivity: "pa02", Amount: 200, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot02-iid01"},
		{DocType: "lot", ID: "lot03", LotType: "test-type", ProdActivity: "pa03", Amount: 300, Unit: "KG", ProdUnit: "punit01", LotInternalID: "lot03-iid01"},
		{DocType: "lot", ID: "lot04", LotType: "test-type", ProdActivity: "pa04", Amount: 400, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot04-iid01"},
		{DocType: "lot", ID: "lot05", LotType: "test-type", ProdActivity: "pa05", Amount: 500, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot05-iid01"},
		{DocType: "lot", ID: "lot06", LotType: "test-type", Amount: 600, Unit: "KG", ProdUnit: "punit02", LotInternalID: "lot06-iid01"},
		{DocType: "lot", ID: "lot07", LotType: "test-type", Amount: 700, Unit: "KG", ProdUnit: "punit03", LotInternalID: "lot07-iid01"},
	}

	prodActivities := []ProdActivity{
		{DocType: "prodActivity", ID: "pa01", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[0], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 73},
		{DocType: "prodActivity", ID: "pa02", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[1], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 16},
		{DocType: "prodActivity", ID: "pa03", ActivityType: "test-type", ProdUnit: "punit01", InputLots: map[string]float32{"lot01": 20, "lot02": 15}, OutputLot: lots[2], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 51},
		{DocType: "prodActivity", ID: "pa04", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[3], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 26},
		{DocType: "prodActivity", ID: "pa05", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot01": 10}, OutputLot: lots[4], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 14},
		{DocType: "prodActivity", ID: "pa06", ActivityType: "test-type", ProdUnit: "punit02", InputLots: map[string]float32{"lot04": 50, "lot05": 20}, OutputLot: lots[5], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 20},
		{DocType: "prodActivity", ID: "pa07", ActivityType: "test-type", ProdUnit: "punit03", InputLots: map[string]float32{"lot01": 30, "lot04": 10, "lot06": 10}, OutputLot: lots[6], ActivityEndDate: "date", CompanyLegalName: "name", Location: "location", EnvScore: 100},
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

	return fmt.Sprintf("production activities [%s-%s] & lots [%s-%s] were successfully added to the ledger", prodActivities[0].ID, prodActivities[len(prodActivities)-1].ID, lots[0].ID, lots[len(lots)-1].ID), nil
}

/*
 * -----------------------------------
 * LOT
 * -----------------------------------
 */

// LotExists returns true when lot with given ID exists in world state
func (c *StvgdContract) LotExists(ctx contractapi.TransactionContextInterface, lotID string) (bool, error) {
	data, err := ctx.GetStub().GetState(lotID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateLot creates a new instance of Lot
func (c *StvgdContract) CreateLot(ctx contractapi.TransactionContextInterface, lotID, lotType, prodActivity string,
	amount float32, unit, prodUnit, lotInternalID string) (string, error) {

	exists, err := c.LotExists(ctx, lotID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("the lot %s already exists", lotID)
	}

	if amount < 0 {
		return "", fmt.Errorf("the amount should be greater than 0")
	} else {
		lot := &Lot{
			DocType:       "lot",
			ID:            lotID,
			LotType:       lotType,
			ProdActivity:  prodActivity,
			Amount:        amount,
			Unit:          unit,
			ProdUnit:      prodUnit,
			LotInternalID: lotInternalID,
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

	return fmt.Sprintf("%s created successfully", lotID), nil
}

// ReadLot retrieves an instance of Lot from the world state
func (c *StvgdContract) ReadLot(ctx contractapi.TransactionContextInterface, lotID string) (*Lot, error) {

	exists, err := c.LotExists(ctx, lotID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the lot %s does not exist", lotID)
	}

	lotBytes, _ := ctx.GetStub().GetState(lotID)

	lot := new(Lot)

	err = json.Unmarshal(lotBytes, lot)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Lot")
	}

	return lot, nil
}

// constructQueryResponseFromIterator constructs a slice of lots from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Lot, error) {
	var lots []*Lot
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var lot Lot
		err = json.Unmarshal(queryResult.Value, &lot)
		if err != nil {
			return nil, err
		}
		lots = append(lots, &lot)
	}

	return lots, nil
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Lot, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// GetAllLots queries for all lots.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (docType).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (c *StvgdContract) GetAllLots(ctx contractapi.TransactionContextInterface) ([]*Lot, error) {
	queryString := `{"selector":{"docType":"lot"}}`
	return getQueryResultForQueryString(ctx, queryString)
}

// UpdateLotAmount updates the amount of a Lot from the world state
func (c *StvgdContract) UpdateLotAmount(ctx contractapi.TransactionContextInterface, lotID string, newAmount float32) (string, error) {

	// Verifies if Lot that has lotID already exists
	exists, err := c.LotExists(ctx, lotID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("the lot %s does not exist", lotID)
	}

	outdatedLotBytes, _ := ctx.GetStub().GetState(lotID) // Gets "old" Lot bytes from lotID

	outdatedLot := new(Lot) // Initialize outdated/"old" Lot object

	// Parses the JSON-encoded data in bytes (outdatedLotBytes) to the "old" Lot object (outdatedLot)
	err = json.Unmarshal(outdatedLotBytes, outdatedLot)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Lot")
	}

	// Checks if amount >= 0
	if newAmount < 0 {
		return "", fmt.Errorf("the new amount should be greater than 0")
	} else {
		// Initialize updated/"new" Lot object
		updatedLot := &Lot{
			DocType:       outdatedLot.DocType,
			ID:            outdatedLot.ID,
			LotType:       outdatedLot.LotType,
			ProdActivity:  outdatedLot.ProdActivity,
			Amount:        newAmount,
			Unit:          outdatedLot.Unit,
			ProdUnit:      outdatedLot.ProdUnit,
			LotInternalID: outdatedLot.LotInternalID,
		}

		updatedLotBytes, _ := json.Marshal(updatedLot) // Encodes the JSON updatedLot data to bytes

		err = ctx.GetStub().PutState(lotID, updatedLotBytes) // Updates world state with newly updated Lot
		if err != nil {
			return "", fmt.Errorf("could not write from world state. %s", err)
		} else if newAmount == 0 { // Deletes the lot if there is no more amount left / newAmount = 0
			_, err = c.DeleteLot(ctx, lotID)
			if err != nil {
				return "", fmt.Errorf("could not delete from world state. %s", err)
			} else {
				return fmt.Sprintf("lot [%s]'s amount was successfully updated to %.2f%s and deleted from world state", lotID, newAmount, outdatedLot.Unit), nil
			}
		} else {
			return fmt.Sprintf("lot [%s]'s amount was successfully updated to %.2f%s", lotID, newAmount, outdatedLot.Unit), nil
		}
	}
}

// TransferLot transfers a lot by setting a new production unit id on the lot
func (c *StvgdContract) TransferLot(ctx contractapi.TransactionContextInterface, lotID, newProdUnit string) (string, error) {

	// Verifies if Lot that has lotID already exists
	exists, err := c.LotExists(ctx, lotID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("the lot %s does not exist", lotID)
	}

	outdatedLotBytes, _ := ctx.GetStub().GetState(lotID) // Gets "old" Lot bytes from lotID

	outdatedLot := new(Lot) // Initialize outdated/"old" Lot object

	// Parses the JSON-encoded data in bytes (outdatedLotBytes) to the "old" Lot object (outdatedLot)
	err = json.Unmarshal(outdatedLotBytes, outdatedLot)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Lot")
	}

	// Checks if new owner is different
	if newProdUnit == outdatedLot.ProdUnit {
		return "", fmt.Errorf("cannot transfer a lot to the current owner/production unit [%s]", outdatedLot.ProdUnit)
	} else {
		// Initialize updated/"new" Lot object
		updatedLot := &Lot{
			DocType:       outdatedLot.DocType,
			ID:            outdatedLot.ID,
			LotType:       outdatedLot.LotType,
			ProdActivity:  outdatedLot.ProdActivity,
			Amount:        outdatedLot.Amount,
			Unit:          outdatedLot.Unit,
			ProdUnit:      newProdUnit,
			LotInternalID: outdatedLot.LotInternalID,
		}

		updatedLotBytes, _ := json.Marshal(updatedLot) // Encodes the JSON updatedLot data to bytes

		err = ctx.GetStub().PutState(lotID, updatedLotBytes) // Updates world state with newly updated Lot
		if err != nil {
			return "", fmt.Errorf("could not write to world state. %s", err)
		} else {
			return fmt.Sprintf("lot [%s] transfered successfully to production unit [%s]", lotID, newProdUnit), nil
		}
	}
}

// DeleteLot deletes an instance of Lot from the world state
func (c *StvgdContract) DeleteLot(ctx contractapi.TransactionContextInterface, lotID string) (string, error) {
	exists, err := c.LotExists(ctx, lotID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("the lot [%s] does not exist", lotID)
	}

	err = ctx.GetStub().DelState(lotID)
	if err != nil {
		return "", fmt.Errorf("could not delete from world state. %s", err)
	} else {
		return fmt.Sprintf("lot [%s] deleted successfully", lotID), nil
	}
}

// DeleteAllLots deletes all lots found in world state
func (c *StvgdContract) DeleteAllLots(ctx contractapi.TransactionContextInterface) (string, error) {

	lots, err := c.GetAllLots(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if len(lots) == 0 {
		return "", fmt.Errorf("there are no lots in world state to delete")
	}

	for _, lot := range lots {
		err = ctx.GetStub().DelState(lot.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete from world state. %s", err)
		}
	}

	return "all the lots were successfully deleted", nil
}

/*
 * -----------------------------------
 * PRODUCTION ACTIVITY
 * -----------------------------------
 */

// ProdActivityExists returns true when prodActivity with given ID exists in world state
func (c *StvgdContract) ProdActivityExists(ctx contractapi.TransactionContextInterface, prodActivityID string) (bool, error) {
	data, err := ctx.GetStub().GetState(prodActivityID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateProdActivity creates a new instance of ProdActivity
func (c *StvgdContract) CreateProdActivity(ctx contractapi.TransactionContextInterface, prodActivityID, activityType, prodUnit string,
	inputLots map[string]float32, outputLot Lot, activityEndDate, companyLegalName, location string, envScore float32) (string, error) {

	// Checks if the output lot ID already exists
	exists, err := c.LotExists(ctx, outputLot.ID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("the lot [%s] already exists", outputLot.ID)
	}

	// Checks if the production activity ID already exists
	exists, err = c.ProdActivityExists(ctx, prodActivityID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("the production activity [%s] already exists", prodActivityID)
	}

	// Checks equality in production activity IDs & production units
	if prodActivityID != outputLot.ProdActivity {
		return "", fmt.Errorf("production activity's ID [%s] must be the same as output lot's production activity's ID [%s]", prodActivityID, outputLot.ProdActivity)
	} else if prodUnit != outputLot.ProdUnit {
		return "", fmt.Errorf("production unit's ID [%s] must be the same as output lot's production unit's ID [%s]", prodUnit, outputLot.ProdUnit)
	}

	// Input lots audit
	if len(inputLots) > 0 { // If production activity uses input lots

		var amountSum float32 = 0 // Local variable to verify if newly created Lot's amount doesn't exceed sum of input lots' amounts

		for lotID, amount := range inputLots { // In every single input lot

			// Checks if the lot ID already exist
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

			// Validate inserted amounts (0 <= amount(inputLot) <= lot.Amount)
			switch {
			case amount <= 0:
				return "", fmt.Errorf("input lots' amounts must be greater than 0 (input amount for lot [%s] is %.2f)", lotID, amount)
			case amount > lot.Amount:
				return "", fmt.Errorf("input lots' amounts must not exceed the lot's total amount (lot [%s] max amount is %.2f)", lotID, lot.Amount)
			}

			amountSum += amount // Increment input lot's amount to sum

			// Subtract lot's amount with input lots' amount //! CURRENTLY NOT WORKING
			_, err = c.UpdateLotAmount(ctx, lotID, lot.Amount-amount)
			if err != nil {
				return "", fmt.Errorf("could not write to world state. %s", err)
			}

			// Transfer input lots ownership to new production unit / owner
			if lot.ProdUnit != prodUnit { // Only transfer is production units for the input lots are different
				_, err = c.TransferLot(ctx, lotID, prodUnit)
				if err != nil {
					return "", fmt.Errorf("could not write to world state. %s", err)
				}
			}
		}

		// Validate output lot's amount (outputLot.Amount > amountSum)
		if outputLot.Amount > amountSum {
			return "", fmt.Errorf("output lot's inserted amount [%.2f] is bigger than the sum of input lots' amounts [%.2f]", outputLot.Amount, amountSum)
		}
	}

	// Create production activity's output lot
	_, err = c.CreateLot(ctx, outputLot.ID, outputLot.LotType, prodActivityID, outputLot.Amount, outputLot.Unit, outputLot.ProdUnit, outputLot.LotInternalID)
	if err != nil {
		return "", fmt.Errorf("could not write to world state. %s", err)
	}

	prodActivity := &ProdActivity{
		DocType:          "prodActivity",
		ID:               prodActivityID,
		ActivityType:     activityType,
		ProdUnit:         prodUnit,
		InputLots:        inputLots,
		OutputLot:        outputLot,
		ActivityEndDate:  activityEndDate,
		CompanyLegalName: companyLegalName,
		Location:         location,
		EnvScore:         envScore,
	}

	prodActivityBytes, err := json.Marshal(prodActivity)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(prodActivityID, prodActivityBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put to world state: %v", err)
	}

	return fmt.Sprintf("production activity [%s] & lot [%s] were successfully added to the ledger", prodActivityID, outputLot.ID), nil
}

// ReadProdActivity retrieves an instance of ProdActivity from the world state
func (c *StvgdContract) ReadProdActivity(ctx contractapi.TransactionContextInterface, prodActivityID string) (*ProdActivity, error) {

	exists, err := c.ProdActivityExists(ctx, prodActivityID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the production activity [%s] does not exist", prodActivityID)
	}

	prodActivityBytes, _ := ctx.GetStub().GetState(prodActivityID)

	prodActivity := new(ProdActivity)

	err = json.Unmarshal(prodActivityBytes, prodActivity)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type ProdActivity")
	}

	return prodActivity, nil
}

// GetAllProdActivities returns all production activities found in world state
func (c *StvgdContract) GetAllProdActivities(ctx contractapi.TransactionContextInterface) ([]*ProdActivity, error) {
	// range query with empty string for endKey does an
	// open-ended query of all production activities in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("pa", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var prodActivities []*ProdActivity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var prodActivity ProdActivity
		err = json.Unmarshal(queryResponse.Value, &prodActivity)
		if err != nil {
			return nil, err
		}
		prodActivities = append(prodActivities, &prodActivity)
	}

	return prodActivities, nil
}

// DeleteProdActivities deletes all production activities found in world state
func (c *StvgdContract) DeleteAllProdActivities(ctx contractapi.TransactionContextInterface) (string, error) {

	prodActivities, err := c.GetAllProdActivities(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if len(prodActivities) == 0 {
		return "", fmt.Errorf("there are no productions activites in world state to delete")
	}

	for _, prodActivity := range prodActivities {
		err = ctx.GetStub().DelState(prodActivity.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete from world state. %s", err)
		}
	}

	return "all the production activities were successfully deleted", nil
}
