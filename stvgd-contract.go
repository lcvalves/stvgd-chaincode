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
