/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const getStateError = "world state get error"

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	args := ms.Called(key)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(key string, value []byte) error {
	args := ms.Called(key, value)

	return args.Error(0)
}

func (ms *MockStub) DelState(key string) error {
	args := ms.Called(key)

	return args.Error(0)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

/*
 * -----------------------------------
 * BATCH
 * -----------------------------------
 */

func configureBatchStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	testBatch := &Batch{
		ObjectType:           "batch",
		ID:                   "batch01",
		BatchTypeID:          "batchType01",
		ProductionActivityID: "productionActivity01",
		ProductionUnitID:     "productionUnit01",
		BatchInternalID:      "batch01",
		SupplierID:           "supplier01",
		BatchComposition:     map[string]float32{"raw_material01": 100},
		Quantity:             150,
		Unit:                 "Kilograms",
		ECS:                  5,
		SES:                  -5,
	}

	lotBytes, _ := json.Marshal(testBatch)

	ms := new(MockStub)
	ms.On("GetState", "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("{\"docType\": \"lot\",\"ID\": \"lot01\",\"lotType\": \"test-type\",\"prodActivity\": \"pa01\",\"amount\": 100,\"unit\": \"KG\",\"prodUnit\": \"punit01\",\"lotInternalID\": \"lot01-iid01\"}"), nil)
	ms.On("GetState", "lotkey").Return(lotBytes, nil)
	ms.On("PutState", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelState", mock.AnythingOfType("string")).Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)

	return mc, ms
}

func TestBatchExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureBatchStub()
	c := new(StvgdContract)

	exists, err = c.BatchExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	exists, err = c.BatchExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = c.BatchExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

/*
func TestCreateBatch(t *testing.T) {
	var err error

	ctx, stub := configureBatchStub()
	c := new(StvgdContract)

	_, err = c.CreateBatch(ctx, "statebad", "test-type", "", 100, "KG", "punit01", "lot01-iid01")
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors")

	_, err = c.CreateBatch(ctx, "existingkey", "test-type", "", 100, "KG", "punit01", "lot01-iid01")
	assert.EqualError(t, err, "the lot existingkey already exists", "should error when exists returns true")

	_, _ = c.CreateBatch(ctx, "missingkey", "test-type", "", 100, "KG", "punit01", "lot01-iid01")
	stub.AssertCalled(t, "PutState", "missingkey", []byte("{\"docType\":\"lot\",\"ID\":\"missingkey\",\"lotType\":\"test-type\",\"amount\":100,\"unit\":\"KG\",\"prodUnit\":\"punit01\",\"lotInternalID\":\"lot01-iid01\"}"))
}

func TestReadBatch(t *testing.T) {
	var lot *Batch
	var err error

	ctx, _ := configureBatchStub()
	c := new(StvgdContract)

	lot, err = c.ReadBatch(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors when reading")
	assert.Nil(t, lot, "should not return Batch when exists errors when reading")

	lot, err = c.ReadBatch(ctx, "missingkey")
	assert.EqualError(t, err, "the lot missingkey does not exist", "should error when exists returns true when reading")
	assert.Nil(t, lot, "should not return Batch when key does not exist in world state when reading")

	// No need for "existingkey" testing due to method being dependent on Batch type

		//lot, err = c.ReadBatch(ctx, "existingkey")
		//assert.EqualError(t, err, "could not unmarshal world state data to type Batch", "should error when data in key is not Batch")
		//assert.Nil(t, lot, "should not return Batch when data in key is not of type Batch")


	lot, err = c.ReadBatch(ctx, "lotkey")
	expectedBatch := &Batch{
		DocType:         "lot",
		ID:              "lot01",
		BatchType:       "test-type",
		ProdActivity:    "pa01",
		Amount:          100,
		Unit:            "KG",
		ProdUnit:        "punit01",
		BatchInternalID: "lot01-iid01",
	}
	assert.Nil(t, err, "should not return error when Lot exists in world state when reading")
	assert.Equal(t, expectedLot, lot, "should return deserialized Lot from world state")
}
*/
/*
//TODO: TestGetAllLots

func TestUpdateLotAmount(t *testing.T) {
	var err error

	ctx, stub := configureLotStub()
	c := new(StvgdContract)

	_, err = c.UpdateLotAmount(ctx, "statebad", 200)
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors when updating")

	_, err = c.UpdateLotAmount(ctx, "missingkey", 200)
	assert.EqualError(t, err, "the lot missingkey does not exist", "should error when exists returns true when updating")

	_, err = c.UpdateLotAmount(ctx, "lotkey", 200)
	expectedLot := &Lot{
		DocType:       "lot",
		ID:            "lot01",
		LotType:       "test-type",
		ProdActivity:  "pa01",
		Amount:        200,
		Unit:          "KG",
		ProdUnit:      "punit01",
		LotInternalID: "lot01-iid01",
	}
	expectedLotBytes, _ := json.Marshal(expectedLot)
	assert.Nil(t, err, "should not return error when Lot exists in world state when updating")
	stub.AssertCalled(t, "PutState", "lotkey", expectedLotBytes)
}

//TODO: TestUpdateLotAmountToZero

func TestDeleteLot(t *testing.T) {
	var err error

	ctx, stub := configureLotStub()
	c := new(StvgdContract)

	_, err = c.DeleteLot(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors")

	_, err = c.DeleteLot(ctx, "missingkey")
	assert.EqualError(t, err, "the lot [missingkey] does not exist", "should error when exists returns true when deleting")

	_, err = c.DeleteLot(ctx, "lotkey")
	assert.Nil(t, err, "should not return error when Lot exists in world state when deleting")
	stub.AssertCalled(t, "DelState", "lotkey")
}

//TODO: TestDeleteAllLots

func configureProdActivityStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	testLot := Lot{
		DocType:       "lot",
		ID:            "lot01",
		LotType:       "test-type",
		ProdActivity:  "pa01",
		Amount:        100,
		Unit:          "KG",
		ProdUnit:      "punit01",
		LotInternalID: "lot01-iid01",
	}

	testProdActivity := &ProdActivity{
		DocType:          "prodActivity",
		ID:               "pa01",
		ActivityType:     "test-type",
		ProdUnit:         "punit01",
		InputLots:        map[string]float32{"lot01": 10},
		OutputLot:        testLot,
		ActivityEndDate:  "test-date",
		CompanyLegalName: "test-name",
		Location:         "test-location",
		EnvScore:         20,
	}

	lotBytes, _ := json.Marshal(testProdActivity)

	ms := new(MockStub)
	ms.On("GetState", "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("{\"docType\":\"prodActivity\",\"ID\":\"pa01\",\"activityType\":\"test-type\",\"prodUnit\":\"punit01\",\"inputLots\":{\"lot01\":10},\"outputLot\":{\"docType\": \"lot\",\"ID\":\"lot01\",\"lotType\":\"test-type\",\"prodActivity\":\"pa01\",\"amount\":100,\"unit\":\"KG\",\"prodUnit\":\"punit01\",\"lotInternalID\":\"lot01-iid01\"},\"activityEndDate\":\"test-date\",\"companyLegalName\":\"test-name\",\"location\":\"test-location\",\"envScore\":20}"), nil)
	ms.On("GetState", "lotkey").Return(lotBytes, nil)
	ms.On("PutState", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelState", mock.AnythingOfType("string")).Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)

	return mc, ms
}
*/
/*
 * -----------------------------------
 * PRODUCTION ACTIVITY
 * -----------------------------------
 */
/*
func TestProdActivityExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureProdActivityStub()
	c := new(StvgdContract)

	exists, err = c.ProdActivityExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false on error")

	exists, err = c.ProdActivityExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = c.ProdActivityExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}
*/
//TODO: TestCreateProdActivity

//TODO: TestReadProdActivity

//TODO: TestGetAllProdActivities

//TODO: TestDeleteAllProdActivities
