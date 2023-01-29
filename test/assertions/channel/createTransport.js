// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json().message;

// Aux Transport variables
var transportTypes = ["ROAD", "MARITIME", "AIR", "RAIL", "INTERMODAL"];

if (res == "User with provided token is not enrolled") {
  // Tests if issuer is enrolled
  pm.test("User is enrolled", function () {
    pm.expect(pm.response.code, "Incorrect request status").to.not.eql(403);
    pm.expect(pm.response.text(), "Missing authentication").to.not.include(
      `User with provided token is not enrolled`
    );
  });
} else {
  // Tests if issuer is enrolled
  pm.test("User is enrolled", function () {
    pm.expect(pm.response.code, "Incorrect request status").to.not.eql(403);
    pm.expect(pm.response.text(), "Missing authentication").to.not.include(
      `User with provided token is not enrolled`
    );
  });

  // Gets contract & method from request body
  var contract = req.method.split(":")[0];
  method = req.method.split(":")[1];

  // Tests for valid request structure
  pm.test(
    "Request strucuture is valid (args are array of strings, no missing channel/chaincode/method names & valid transient params)",
    function () {
      pm.expect(pm.response.text(), "Missing channel name").to.not.include(
        `Missing channel name in path`
      );
      pm.expect(pm.response.text(), "Missing chaincode name").to.not.include(
        `Missing chaincode name in path`
      );
      pm.expect(pm.response.text(), "Missing contract name").to.not.include(
        `Contract not found with name ${contract}`
      );
      pm.expect(pm.response.text(), "Missing chaincode method").to.not.include(
        `Missing chaincode method in request body`
      );
      pm.expect(pm.response.text(), "Missing chaincode method").to.not.include(
        `Blank function name passed`
      );
      pm.expect(pm.response.text(), "Missing chaincode method").to.not.include(
        `Function ${method} not found in contract ${contract}`
      );
      pm.expect(pm.response.text(), "Invalid args type").to.not.include(
        `Invalid chaincode args. It must be an array of strings`
      );
      pm.expect(
        pm.response.text(),
        "Invalid transient parameter"
      ).to.not.include(
        `Invalid transient parameter. It must be an object with string keys and string values`
      );
    }
  );

  // Aux variables assignment for data comparison
  var transportID = req.args[0];
  originProductionUnitInternalID = req.args[1];
  destinationProductionUnitID = req.args[2];
  transportType = req.args[3];
  activityDate = req.args[4];
  inputBatches = new Map(Object.entries(JSON.parse(req.args[5])));
  isReturn = req.args[6];

  // Tests for valid transport ID
  pm.test("Valid transport ID", function () {
    pm.expect(transportID, "Empty transport ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `activity ID prefix must match its type (should be [t-...])`
    );
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `incorrect activity prefix`
    );
    pm.expect(pm.response.text(), "Existing activity ID").to.not.include(
      `transport activity [${transportID}] already exists`
    );
  });

  // Tests for valid timestamp
  pm.test("Valid timestamps", function () {
    pm.expect(pm.response.text(), "Invalid timestamp").to.not.include(
      `could not get transaction timestamp:`
    );
    pm.expect(pm.response.text(), "Invalid timestamp").to.not.include(
      `could not parse activity start date:`
    );
    pm.expect(pm.response.text(), "Invalid timestamp").to.not.include(
      `activity start date can't be after the activity end date:`
    );
  });

  // Tests for valid activity type
  pm.test("Transport type is valid (ROAD, MARITIME, AIR, ...)", function () {
    pm.expect(transportType, "Empty activity type").to.not.be.empty;
    pm.expect(
      transportTypes,
      "Inserted activity type not defined"
    ).to.deep.include(transportType);
    pm.expect(pm.response.text(), "Invalid activity type").to.not.include(
      `could not validate activity type:`
    );
  });

  // Tests for valid transport unit internal ID
  pm.test("Production unit IDs are valid (not empty)", function () {
    pm.expect(
      originProductionUnitInternalID,
      "Empty origin production unit internal ID"
    ).to.not.be.empty;
    pm.expect(
      destinationProductionUnitID,
      "Empty destination production unit ID"
    ).to.not.be.empty;
    pm.expect(
      pm.response.text(),
      "Invalid transport unit internal ID"
    ).to.not.include(`destination production unit's ID must not be empty`);
    pm.expect(
      pm.response.text(),
      "Origin can't be the same as destination"
    ).to.not.include(
      `must be different from destination production unit ID [${destinationProductionUnitID}]`
    );
    pm.expect(
      pm.response.text(),
      "Can only transport self-owned batches"
    ).to.not.include(
      `can only transport batches that are in current production unit`
    );
  });

  // Tests for valid company MSP ID
  pm.test("Got MSP ID", function () {
    pm.expect(pm.response.text(), "Invalid MSP ID").to.not.include(
      `could not get MSP ID:`
    );
  });

  // Tests for valid issuer's ID
  pm.test("Got issuer's ID", function () {
    pm.expect(pm.response.text(), "Invalid issuer ID").to.not.include(
      `could not get issuer's client ID:`
    );
  });

  // Tests for valid input batches data
  pm.test("Input batches are valid", function () {
    pm.expect(
      inputBatches.size,
      "Invalid input batches (must have only 1 batch)"
    ).to.be.eql(1);
    inputBatches.forEach((value, key) => {
      pm.expect(pm.response.text(), "Error reading batch").to.not.include(
        `could not read batch from world state:`
      );
      pm.expect(pm.response.text(), "Invalid batch ID").to.not.include(
        `batch [${key}] does not exist`
      );
      pm.expect(pm.response.text(), "Input batch in transit").to.not.include(
        `batch [${key}] currently in transit`
      );
      pm.expect(
        pm.response.text(),
        "Must return total quantity if it is a return transport"
      ).to.not.include(
        `when returning a batch, input batch quantity [${value}] must be equal to batch's total quantity`
      );
      pm.expect(
        value,
        "Invalid input batch quantity (must be positive [+0])"
      ).to.be.above(0);
      pm.expect(
        pm.response.text(),
        "Invalid input batch quantity (exceeds batch total quantity"
      ).to.not.include(
        `input batches' quantities must not exceed the batch's total quantity`
      );
      pm.expect(pm.response.text(), "Error updating batch").to.not.include(
        `failed to put batch to world state:`
      );
      pm.expect(pm.response.text(), "Error deleting batch").to.not.include(
        `failed to delete batch to world state:`
      );
    });
  });

  // Tests for transport and batch creation
  pm.test("Batch & Transport are added to state", function () {
    pm.expect(pm.response.text(), "Error updating batch").to.not.include(
      `failed to put remaining batch to world state:`
    );
    pm.expect(pm.response.text(), "Error adding batch").to.not.include(
      `failed to put updated batch to world state:`
    );
    pm.expect(pm.response.text(), "Error adding transport").to.not.include(
      `failed to put transport to world state:`
    );
    pm.expect(pm.response.text(), "Added transport & batch").to.include(
      `successfully added to the ledger`
    );
  });
}
