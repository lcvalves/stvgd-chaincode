package main

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/civil"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/*
 * -----------------------------------
 * STRUCTS
 * -----------------------------------
 */

// Registration stores information about the batch receptions in the supply chain companies/production units
type Registration struct {
	ObjectType        string         `json:"objType"` // objType ("rg") is used to distinguish the various types of objects in state database
	ID                string         `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID  string         `json:"productionUnitID"`
	NewBatch          Batch          `json:"newBatch"`
	ActivityStartDate civil.DateTime `json:"activityStartDate"`
	ActivityEndDate   civil.DateTime `json:"activityEndDate"`
}

/*
 * -----------------------------------
 * TRANSACTIONS
 * -----------------------------------
 */

// RegistrationExists returns true when registration with given ID exists in world state
func (c *StvgdContract) RegistrationExists(ctx contractapi.TransactionContextInterface, registrationID string) (bool, error) {

	// Searches for any world state data under the given registrationID
	data, err := ctx.GetStub().GetState(registrationID)
	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateRegistration creates a new instance of Registration
func (c *StvgdContract) CreateRegistration(ctx contractapi.TransactionContextInterface, registrationID, productionUnitID, activityStartDate, activityEndDate string, newBatch Batch) (string, error) {

	// Checks if the registrationy ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return "", fmt.Errorf("registration [%s] already exists", registrationID)
	}

	// Validate dates
	civilDates, err := validateDates(activityStartDate, activityEndDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates. %s", err)
	}

	// Checks equality in production unit IDs
	if productionUnitID != newBatch.ProductionUnitID {
		return "", fmt.Errorf("production unit's ID [%s] must be the same as output batch's production unit's ID [%s]", productionUnitID, newBatch.ProductionUnitID)
	}

	// Validate new batch
	validBatch, err := validateBatch(ctx, newBatch.ID, newBatch.ProductionUnitID, newBatch.BatchInternalID, newBatch.SupplierID, string(newBatch.Unit), string(newBatch.BatchTypeID), newBatch.BatchComposition, newBatch.Quantity, newBatch.ECS, newBatch.SES)
	if !validBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %v", err)
	}

	// Setup Traceability
	activities := make([]string, 0)
	activities = append(activities, registrationID)
	newBatch.Traceability = Traceability{
		Activities:    activities,
		ParentBatches: make([]string, 0),
	}

	// Instatiate registration
	registration := &Registration{
		ObjectType:        "rg",
		ID:                registrationID,
		ProductionUnitID:  productionUnitID,
		NewBatch:          newBatch,
		ActivityStartDate: civilDates[0],
		ActivityEndDate:   civilDates[1],
	}

	// Marshal batch to bytes
	batchBytes, err := json.Marshal(newBatch)
	if err != nil {
		return "", err
	}
	// Put batchBytes in world state
	err = ctx.GetStub().PutState(newBatch.ID, batchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put batch to world state: %v", err)
	}
	// Marshal registration to bytes
	registrationBytes, err := json.Marshal(registration)
	if err != nil {
		return "", err
	}
	// Put registrationBytes in world state
	err = ctx.GetStub().PutState(registrationID, registrationBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put registration to world state: %v", err)
	}

	return fmt.Sprintf("registration [%s] & batch [%s] were successfully added to the ledger", registrationID, newBatch.ID), nil

}

// ReadRegistration retrieves an instance of Registration from the world state
func (c *StvgdContract) ReadRegistration(ctx contractapi.TransactionContextInterface, registrationID string) (*Registration, error) {

	// Checks if the registration ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("could not read registration from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("registration [%s] does not exist", registrationID)
	}

	// Queries world state for registration with given ID
	registrationBytes, _ := ctx.GetStub().GetState(registrationID)
	// Instatiate registration
	registration := new(Registration)
	// Unmarshal registrationBytes to JSON
	err = json.Unmarshal(registrationBytes, registration)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Registration")
	}

	return registration, nil
}

// GetAllRegistrations returns all registrations found in world state
func (c *StvgdContract) GetAllRegistrations(ctx contractapi.TransactionContextInterface) ([]*Registration, error) {
	// range query with empty string for endKey does an open-ended query of all registrations in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("rg", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var registrations []*Registration
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var registration Registration
		err = json.Unmarshal(queryResponse.Value, &registration)
		if err != nil {
			return nil, err
		}
		registrations = append(registrations, &registration)
	}

	return registrations, nil
}

// DeleteRegistration deletes an instance of Registration from the world state
func (c *StvgdContract) DeleteRegistration(ctx contractapi.TransactionContextInterface, registrationID string) (string, error) {

	// Checks if the registration ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return "", fmt.Errorf("could not read registration from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("registration [%s] does not exist", registrationID)
	}

	// Deletes registration in the world state
	err = ctx.GetStub().DelState(registrationID)
	if err != nil {
		return "", fmt.Errorf("could not delete registration from world state. %s", err)
	} else {
		return fmt.Sprintf("registration [%s] deleted successfully", registrationID), nil
	}
}

// DeleteAllRegistrations deletes all registrations found in world state
func (c *StvgdContract) DeleteAllRegistrations(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the registrations in world state
	registrations, err := c.GetAllRegistrations(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read registrations from world state. %s", err)
	} else if len(registrations) == 0 {
		return "", fmt.Errorf("there are no registrations in world state to delete")
	}

	// Iterate through registrations slice
	for _, registration := range registrations {
		// Delete each registration from world state
		err = ctx.GetStub().DelState(registration.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete registrations from world state. %s", err)
		}
	}

	return "all the registrations were successfully deleted", nil
}
