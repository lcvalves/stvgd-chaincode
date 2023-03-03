package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/lcvalves/stvgd-chaincode/pkg/domain"
)

// StvgdContract contract for managing CRUD for STVgoDigital value chain operations
type StvgdContract struct {
	contractapi.Contract
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *domain.Batch `json:"record"`
	TxId      string        `json:"txId"`
	Timestamp time.Time     `json:"timestamp"`
	IsDelete  bool          `json:"isDelete"`
}

/*
 * -----------------------------------
 * AUX Functions
 * -----------------------------------
 */

// Score validation
func validateScore(score float32) (bool, error) {
	// Range validate score
	if score < -10.0 || score > 10.0 {
		return false, fmt.Errorf("invalid score")
	}
	return true, nil
}

// TxTimestamp RFC3339 Time formatting
func getTxTimestampRFC3339Time(stub shim.ChaincodeStubInterface) (time.Time, error) {
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return time.Now(), err
	}
	tm := time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
	//return tm.Format(time.RFC3339), nil
	return tm, nil
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func getSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

func iterate(data interface{}) interface{} {
	d := reflect.ValueOf(data)
	if reflect.ValueOf(data).Kind() == reflect.Slice {
		returnSlice := make([]interface{}, d.Len())
		for i := 0; i < d.Len(); i++ {
			returnSlice[i] = iterate(d.Index(i).Interface())
		}
		return returnSlice
	} else if reflect.ValueOf(data).Kind() == reflect.Map {
		tmpData := make(map[string]interface{})
		for _, k := range d.MapKeys() {
			tmpData[k.String()] = iterate(d.MapIndex(k).Interface())
		}
		return tmpData
	} else {
		return data
	}
}

/*
func iterate(data interface{}, field string) interface{} {

	// reads value of input data
	d := reflect.ValueOf(data)

	// based on kind
	switch reflect.ValueOf(data).Kind() {
	case reflect.Slice: // if slice

		returnSlice := make([]interface{}, d.Len())

		// iterate through slice
		for i := 0; i < d.Len(); i++ {

			// calls iterate on slice's first element tracebility[0]
			returnSlice[i] = iterate(d.Index(i).Interface(), field)
		}

		return returnSlice

	case reflect.Map: // if map

		tmpData := make(map[string]interface{})

		// if field is not key in map
		if d.FieldByName(field).IsZero() {

			// iterate through map keys
			for _, v := range d.MapKeys() {

				// calls iterate on maps's first key: docType = "t"
				tmpData[v.String()] = iterate(d.MapIndex(v).Interface(), field)
			}
			return tmpData
		} else {
			return d.FieldByName(field)
		}
	}
	return data
}
*/
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
func validateProductionType(productionTypeID string) (domain.ProductionType, error) {
	var productionType domain.ProductionType
	switch productionTypeID {
	case "SPINNING":
		productionType = domain.Spinning
	case "WEAVING":
		productionType = domain.Weaving
	case "KNITTING":
		productionType = domain.Knitting
	case "DYEING_FINISHING":
		productionType = domain.DyeingFinishing
	case "CONFECTION":
		productionType = domain.Confection
	default:
		return "", fmt.Errorf("production type not found")
	}

	return productionType, nil
}

// Transport type validation
func validateTransportType(transportTypeID string) (domain.TransportType, error) {
	var transportType domain.TransportType
	switch transportTypeID {
	case "TERRESTRIAL_SMALL":
		transportType = domain.TerrestrialSmall
	case "TERRESTRIAL_BIG":
		transportType = domain.TerrestrialBig
	case "MARITIME":
		transportType = domain.Maritime
	case "AERIAL":
		transportType = domain.Aerial
	case "RAILROADER":
		transportType = domain.Railroader
	default:
		return "", fmt.Errorf("transport type not found")
	}

	return transportType, nil
}

/*
 * -----------------------------------
 - BATCH Validation
 * -----------------------------------
*/

// Batch type validation
func validateBatchType(batchTypeID string) (domain.BatchType, error) {
	var batchType domain.BatchType
	switch batchTypeID {
	case "CONVENTIONAL_COTTON":
		batchType = domain.ConventionalCotton
	case "ORGANIC_COTTON":
		batchType = domain.OrganicCotton
	case "RECYCLED_COTTON":
		batchType = domain.RecycledCotton
	case "PES":
		batchType = domain.Pes
	case "PES_RPET":
		batchType = domain.PesRPet
	case "POLYPROPYLENE":
		batchType = domain.Polypropylene
	case "POLYAMIDE_6":
		batchType = domain.Polyamide6
	case "POLYAMIDE_66":
		batchType = domain.Polyamide66
	case "PAN":
		batchType = domain.Pan
	case "VISCOSE":
		batchType = domain.Viscose
	case "FLAX":
		batchType = domain.Flax
	case "JUTE":
		batchType = domain.Jute
	case "KENAF":
		batchType = domain.Kenaf
	case "BAMBOO":
		batchType = domain.Bamboo
	case "SILK":
		batchType = domain.Silk
	case "WOOL":
		batchType = domain.Wool
	case "ELASTANE":
		batchType = domain.Elastane
	case "YARN":
		batchType = domain.Yarn
	case "RAW_FABRIC":
		batchType = domain.RawFabric
	case "DYED_FABRIC":
		batchType = domain.DyedFabric
	case "RAW_KNITTED_FABRIC":
		batchType = domain.RawKnittedFabric
	case "DYED_KNITTED_FABRIC":
		batchType = domain.DyedKnittedFabric
	case "GARMENT":
		batchType = domain.Garment
	default:
		return "", fmt.Errorf("batch type not found")
	}

	return batchType, nil
}

// Unit validation
func validateUnit(unitID string) (domain.Unit, error) {
	var unit domain.Unit
	switch unitID {
	case "KG":
		unit = domain.Kilograms
	case "L":
		unit = domain.Liters
	case "M":
		unit = domain.Meters
	case "M2":
		unit = domain.SquaredMeters
	default:
		return "", fmt.Errorf("unit not found")
	}

	return unit, nil
}

// validateBatch validates batch for correct inputs/fields on Registration & Production activities
func validateBatch(ctx contractapi.TransactionContextInterface, batchID, productionUnitID, batchInternalID, supplierID, unit, batchType string, batchComposition map[string]float32, quantity, finalScore float32, isInTransit bool) (bool, error) {

	/// Batch prefix validation
	if batchID == "" {
		return false, fmt.Errorf("incorrect batch prefix. (should be [b-...])")
	}
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
	case "CONVENTIONAL_COTTON":
	case "ORGANIC_COTTON":
	case "RECYCLED_COTTON":
	case "PES":
	case "PES_RPET":
	case "POLYPROPYLENE":
	case "POLYAMIDE_6":
	case "POLYAMIDE_66":
	case "PAN":
	case "VISCOSE":
	case "FLAX":
	case "JUTE":
	case "KENAF":
	case "BAMBOO":
	case "SILK":
	case "WOOL":
	case "ELASTANE":
	case "YARN":
	case "RAW_FABRIC":
	case "DYED_FABRIC":
	case "RAW_KNITTED_FABRIC":
	case "DYED_KNITTED_FABRIC":
	case "GARMENT":
	default:
		return false, fmt.Errorf("batch type is not valid")
	}

	// Validate batch composition
	var percentageSum float32 = 0.00 // Local variable for percentage sum validation
	for _, percentage := range batchComposition {
		percentageSum += percentage
		if percentageSum > 100 {
			return false, fmt.Errorf("batch composition percentage sum should be equal to 100")
		}
	}
	if percentageSum != 100 {
		return false, fmt.Errorf("batch composition percentage sum should be equal to 100")
	}

	if quantity < 0 || fmt.Sprintf("%f", quantity) == "" {
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

	// Validate scores (-10 <= SCORE <= 10)
	validScores, err := validateScore(finalScore)
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
func getQueryResultForQueryStringBatch(ctx contractapi.TransactionContextInterface, queryString string) ([]*domain.Batch, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorBatch(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of batches from the resultsIterator
func constructQueryResponseFromIteratorBatch(resultsIterator shim.StateQueryIteratorInterface) ([]*domain.Batch, error) {
	var batches []*domain.Batch
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var batch domain.Batch
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
func getQueryResultForQueryStringRegistration(ctx contractapi.TransactionContextInterface, queryString string) ([]*domain.Registration, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorRegistration(resultsIterator)
}

// constructQueryResponseFromIteratorRegistration constructs a slice of registrations from the resultsIterator
func constructQueryResponseFromIteratorRegistration(resultsIterator shim.StateQueryIteratorInterface) ([]*domain.Registration, error) {
	var registrations []*domain.Registration
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var registration domain.Registration
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
func getQueryResultForQueryStringProduction(ctx contractapi.TransactionContextInterface, queryString string) ([]*domain.Production, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorProduction(resultsIterator)
}

// constructQueryResponseFromIteratorProduction constructs a slice of production activities from the resultsIterator
func constructQueryResponseFromIteratorProduction(resultsIterator shim.StateQueryIteratorInterface) ([]*domain.Production, error) {
	var productions []*domain.Production
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var production domain.Production
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
func getQueryResultForQueryStringTransport(ctx contractapi.TransactionContextInterface, queryString string) ([]*domain.Transport, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorTransport(resultsIterator)
}

// constructQueryResponseFromIteratorTransport constructs a slice of batches from the resultsIterator
func constructQueryResponseFromIteratorTransport(resultsIterator shim.StateQueryIteratorInterface) ([]*domain.Transport, error) {
	var transports []*domain.Transport
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transport domain.Transport
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
func getQueryResultForQueryStringReception(ctx contractapi.TransactionContextInterface, queryString string) ([]*domain.Reception, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIteratorReception(resultsIterator)
}

// constructQueryResponseFromIteratorReception constructs a slice of receptions from the resultsIterator
func constructQueryResponseFromIteratorReception(resultsIterator shim.StateQueryIteratorInterface) ([]*domain.Reception, error) {
	var receptions []*domain.Reception
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var reception domain.Reception
		err = json.Unmarshal(queryResult.Value, &reception)
		if err != nil {
			return nil, err
		}
		receptions = append(receptions, &reception)
	}

	return receptions, nil
}
