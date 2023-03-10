package app

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/lcvalves/stvgd-chaincode/pkg/domain"
)

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
func (c *StvgdContract) CreateRegistration(ctx contractapi.TransactionContextInterface, registrationID, productionUnitInternalID, batchID, batchType, batchInternalID, supplierID string, quantity, finalScore float32, batchComposition map[string]float32) (string, error) {

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

	// Timestamp when the transaction was created, have the same value across all endorsers
	txTimestamp, err := getTxTimestampRFC3339Time(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("could not get transaction timestamp: %w", err)
	}

	// Validate batch type
	validBatchType, err := validateBatchType(batchType)
	if err != nil {
		return "", fmt.Errorf("could not validate batch type %w", err)
	}

	/// Validate production unit internal ID
	if productionUnitInternalID == "" {
		return "", fmt.Errorf("production unit internal ID must not be empty: %w", err)
	}

	/// Get company MSP ID
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not get MSP ID: %w", err)
	}

	// Get issuer client ID
	clientID, err := getSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get issuer's client ID: %w", err)
	}

	if quantity <= 0 {
		return "", fmt.Errorf("batch quantity should be positive")
	}

	// Range validate score
	if fmt.Sprintf("%f", finalScore) == "" {
		return "", fmt.Errorf("invalid score")
	}

	// Initialize "new" Batch object
	newBatch := &domain.Batch{
		DocType:          "b",
		ID:               batchID,
		BatchType:        validBatchType,
		LatestOwner:      mspID + ":" + productionUnitInternalID,
		BatchInternalID:  batchInternalID,
		SupplierID:       supplierID,
		Quantity:         quantity,
		FinalScore:       finalScore,
		BatchComposition: batchComposition,
	}

	// Validate new batch
	isValidBatch, err := validateBatch(ctx, newBatch.ID, newBatch.LatestOwner, newBatch.BatchInternalID, newBatch.SupplierID, string(newBatch.BatchType), newBatch.BatchComposition, newBatch.Quantity, newBatch.FinalScore, newBatch.IsInTransit)
	if !isValidBatch {
		return "", fmt.Errorf("failed to validate batch to world state: %w", err)
	}

	// Instatiate registration
	registration := &domain.Registration{
		DocType:          "rg",
		ID:               registrationID,
		ProductionUnitID: newBatch.LatestOwner,
		Issuer:           clientID,
		NewBatch:         *newBatch,
		ActivityDate:     txTimestamp,
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
func (c *StvgdContract) ReadRegistration(ctx contractapi.TransactionContextInterface, registrationID string) (*domain.Registration, error) {

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
	registration := new(domain.Registration)
	// Unmarshal registrationBytes to JSON
	err = json.Unmarshal(registrationBytes, registration)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Registration: %w", err)
	}

	return registration, nil
}

// GetAllRegistrations returns all registrations found in world state
func (c *StvgdContract) GetAllRegistrations(ctx contractapi.TransactionContextInterface) ([]*domain.Registration, error) {
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
