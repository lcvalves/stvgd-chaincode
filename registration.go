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

// Registration stores information about the batch registrations in the supply chain companies/production units
type Registration struct {
	DocType          string         `json:"docType"` // docType ("rg") is used to distinguish the various types of objects in state database
	ID               string         `json:"ID"`      // the field tags are needed to keep case from bouncing around
	ProductionUnitID string         `json:"productionUnitID"`
	ActivityDate     civil.DateTime `json:"activityDate"`
	NewBatch         Batch          `json:"newBatch"`
	ECS              float32        `json:"ecs"` // from non-recorded activities to 1st current owner on the value chain
	SES              float32        `json:"ses"` // from non-recorded activities to 1st current owner on the value chain
}

/*
 * -----------------------------------
 * TRANSACTIONS / METHODS
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
func (c *StvgdContract) CreateRegistration(ctx contractapi.TransactionContextInterface, registrationID, activityDate string, ECS, SES float32, newBatch Batch) (string, error) {

	// Activity prefix validation
	activityPrefix, err := validateActivityType(registrationID)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	} else if activityPrefix != "rg" {
		return "", fmt.Errorf("activity ID prefix must match its type (should be [rg-...])")
	}

	// Checks if the registration ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state: %w", err)
	} else if exists {
		return "", fmt.Errorf("registration [%s] already exists", registrationID)
	}

	// Validate dates
	civilDate, err := civil.ParseDateTime(activityDate)
	if err != nil {
		return "", fmt.Errorf("could not validate dates: %w", err)
	}

	// Validate scores
	validScores, err := validateScores(ECS, SES)
	if !validScores {
		return "", fmt.Errorf("invalid scores: %w", err)
	}

	// Validate new batch
	isValidBatch, err := validateBatch(ctx, newBatch.ID, newBatch.ProductionUnitID, newBatch.BatchInternalID, newBatch.SupplierID, string(newBatch.Unit), string(newBatch.BatchType), newBatch.BatchComposition, newBatch.Quantity, newBatch.ECS, newBatch.SES, newBatch.IsInTransit)
	if !isValidBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %w", err)
	}

	// Instatiate registration
	registration := &Registration{
		DocType:          "rg",
		ID:               registrationID,
		ProductionUnitID: newBatch.ProductionUnitID,
		NewBatch:         newBatch,
		ActivityDate:     civilDate,
		ECS:              ECS,
		SES:              SES,
	}

	// Setup & append traceability to new batch
	newBatch.Traceability = append(newBatch.Traceability, registration)

	// Marshal new batch to bytes
	batchBytes, err := json.Marshal(newBatch)
	if err != nil {
		return "", err
	}
	// Put batchBytes in world state
	err = ctx.GetStub().PutState(newBatch.ID, batchBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put batch to world state: %w", err)
	}

	// Marshal registration to bytes
	registrationBytes, err := json.Marshal(registration)
	if err != nil {
		return "", err
	}
	// Put registrationBytes in world state
	err = ctx.GetStub().PutState(registrationID, registrationBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put registration to world state: %w", err)
	}

	return fmt.Sprintf("registration [%s] & batch [%s] were successfully added to the ledger", registrationID, newBatch.ID), nil
}

// ReadRegistration retrieves an instance of Registration from the world state
func (c *StvgdContract) ReadRegistration(ctx contractapi.TransactionContextInterface, registrationID string) (*Registration, error) {

	// Checks if the registration ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("could not read registration from world state: %w", err)
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
		return nil, fmt.Errorf("could not unmarshal world state data to type Registration: %w", err)
	}

	return registration, nil
}

// GetAllRegistrations returns all registrations found in world state
func (c *StvgdContract) GetAllRegistrations(ctx contractapi.TransactionContextInterface) ([]*Registration, error) {
	queryString := `{"selector":{"docType":"rg"}}`
	return getQueryResultForQueryStringRegistration(ctx, queryString)
}

// DeleteRegistration deletes an instance of Registration from the world state
func (c *StvgdContract) DeleteRegistration(ctx contractapi.TransactionContextInterface, registrationID string) (string, error) {

	// Checks if the registration ID already exists
	exists, err := c.RegistrationExists(ctx, registrationID)
	if err != nil {
		return "", fmt.Errorf("could not read registration from world state: %w", err)
	} else if !exists {
		return "", fmt.Errorf("registration [%s] does not exist", registrationID)
	}

	// Deletes registration in the world state
	err = ctx.GetStub().DelState(registrationID)
	if err != nil {
		return "", fmt.Errorf("could not delete registration from world state: %w", err)
	} else {
		return fmt.Sprintf("registration [%s] deleted successfully", registrationID), nil
	}
}

// DeleteAllRegistrations deletes all registrations found in world state
func (c *StvgdContract) DeleteAllRegistrations(ctx contractapi.TransactionContextInterface) (string, error) {

	// Gets all the registrations in world state
	registrations, err := c.GetAllRegistrations(ctx)
	if err != nil {
		return "", fmt.Errorf("could not read registrations from world state: %w", err)
	} else if len(registrations) == 0 {
		return "", fmt.Errorf("there are no registrations in world state to delete")
	}

	// Iterate through registrations slice
	for _, registration := range registrations {
		// Delete each registration from world state
		err = ctx.GetStub().DelState(registration.ID)
		if err != nil {
			return "", fmt.Errorf("could not delete registrations from world state: %w", err)
		}
	}

	return "all the registrations were successfully deleted", nil
}
