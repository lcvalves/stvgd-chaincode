/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// StvgdContract contract for managing CRUD for Stvgd
type StvgdContract struct {
	contractapi.Contract
}

// LotExists returns true when lot with given ID exists in world state
func (c *StvgdContract) LotExists(ctx contractapi.TransactionContextInterface, lotID string) (bool, error) {
	data, err := ctx.GetStub().GetState(lotID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// InitLedgerLot adds a base set of lots to the ledger
func (c *StvgdContract) InitLedgerLot(ctx contractapi.TransactionContextInterface) (string, error) {
	lots := []Lot{
		{DocType: "lot", ID: "lot01", LotType: "test-type", ProdActivity: "pa01", Amount: 100, Unit: "0", ProdUnit: "punit01", LotInternalID: "lot01-iid01"},
		{DocType: "lot", ID: "lot02", LotType: "test-type", ProdActivity: "pa02", Amount: 200, Unit: "0", ProdUnit: "punit01", LotInternalID: "lot02-iid01"},
		{DocType: "lot", ID: "lot03", LotType: "test-type", ProdActivity: "pa03", Amount: 300, Unit: "1", ProdUnit: "punit01", LotInternalID: "lot03-iid01"},
		{DocType: "lot", ID: "lot04", LotType: "test-type", ProdActivity: "pa04", Amount: 400, Unit: "0", ProdUnit: "punit02", LotInternalID: "lot04-iid01"},
		{DocType: "lot", ID: "lot05", LotType: "test-type", ProdActivity: "pa05", Amount: 500, Unit: "1", ProdUnit: "punit02", LotInternalID: "lot05-iid01"},
		{DocType: "lot", ID: "lot06", LotType: "test-type", Amount: 600, Unit: "0", ProdUnit: "punit02", LotInternalID: "lot06-iid01"},
		{DocType: "lot", ID: "lot07", LotType: "test-type", Amount: 700, Unit: "1", ProdUnit: "punit03", LotInternalID: "lot07-iid01"},
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

	return "lots [" + lots[0].ID + "-" + lots[len(lots)-1].ID + "] were successfully added to the ledger", nil
}
