// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json().message;

// Aux Production variables
var batchTypes = [
  "FIBER",
  "YARN",
  "MESH",
  "FABRIC",
  "DYED_MESH",
  "FINISHED_MESH",
  "DYED_FABRIC",
  "FINISHED_FABRIC",
  "CUT",
  "FINISHED_PIECE",
  "OTHER",
];
units = ["KG", "L", "M", "M2"];
productionTypes = [
  "SPINNING",
  "WEAVING",
  "KNITTING",
  "DYEING_FINISHING",
  "CONFECTION",
];
compositionSum = 0;

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
  var productionID = req.args[0];
  productionUnitInternalID = req.args[1];
  productionType = req.args[2];
  activityStartDate = req.args[3];
  batchID = req.args[4];
  batchType = req.args[5];
  batchInternalID = req.args[6];
  supplierID = req.args[7];
  unit = req.args[8];
  inputBatches = new Map(Object.entries(JSON.parse(req.args[9])));
  batchComposition = new Map(Object.entries(JSON.parse(req.args[10])));
  quantity = req.args[11];
  finalScore = req.args[12];
  productionScore = req.args[13];
  ses = req.args[14];

  // Tests for valid production ID
  pm.test("Valid production ID", function () {
    pm.expect(productionID, "Empty production ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `activity ID prefix must match its type (should be [p-...])`
    );
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `incorrect activity prefix`
    );
    pm.expect(pm.response.text(), "Existing activity ID").to.not.include(
      `production [${productionID}] already exists`
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
  pm.test(
    "Production type is valid (SPINNING, WEAVING, KNITTING, ...)",
    function () {
      pm.expect(productionType, "Empty activity type ID").to.not.be.empty;
      pm.expect(
        productionTypes,
        "Inserted activity type not defined"
      ).to.deep.include(productionType);
      pm.expect(pm.response.text(), "Invalid activity type").to.not.include(
        `could not validate activity type:`
      );
    }
  );

  // Tests for valid production unit internal ID
  pm.test("Production unit internal ID is valid (not empty)", function () {
    pm.expect(
      productionUnitInternalID,
      "Empty production unit intenral ID"
    ).to.not.be.empty;
    pm.expect(
      pm.response.text(),
      "Invalid production unit internal ID"
    ).to.not.include(`production unit internal ID must not be empty:`);
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

  // Tests for valid scores
  pm.test("Valid activity scores", function () {
    pm.expect(
      parseFloat(productionScore),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
    pm.expect(
      parseFloat(ses),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
  });

  // Tests for valid input batches data
  pm.test("Input batches are valid", function () {
    pm.expect(
      inputBatches.size,
      "Invalid input batches (must have at least 1 batch)"
    ).to.be.at.least(1);
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

  // Tests for valid output batch data
  pm.test("Batch data is valid", function () {
    // Batch ID
    pm.expect(batchID, "Empty batch ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Error reading batch").to.not.include(
      `could not read batch from world state:`
    );
    pm.expect(pm.response.text(), "Invalid batch ID").to.not.include(
      `batch [${batchID}] already exists`
    );
    pm.expect(pm.response.text(), "Invalid batch ID prefix").to.not.include(
      `incorrect batch prefix. (should be [b-...])`
    );

    // Batch internal ID
    pm.expect(batchInternalID, "Empty batch internal ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid batch internal ID").to.not.include(
      `batch internal ID must not be empty`
    );

    // Unit
    pm.expect(unit, "Empty unit").to.not.be.empty;
    pm.expect(units, "Inserted unit not defined").to.be.include(unit);
    pm.expect(pm.response.text(), "Invalid batch unit").to.not.include(
      `could not validate batch unit`
    );

    // Supplier ID
    pm.expect(supplierID, "Empty supplier ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid supplier ID").to.not.include(
      `supplier ID must not be empty`
    );

    // Batch type
    pm.expect(batchType, "Empty batch type ID").to.not.be.empty;
    pm.expect(batchTypes, "Inserted batch type not defined").to.deep.include(
      batchType
    );
    pm.expect(pm.response.text(), "Invalid batch type").to.not.include(
      `could not validate batch type`
    );

    // Quantity
    pm.expect(parseFloat(quantity), "Invalid quantity").to.be.above(0);

    // Batch composition
    pm.expect(
      batchComposition.size,
      "Invalid batch composition (must have at least 1 material)"
    ).to.be.at.least(1);
    batchComposition.forEach((value, key) => {
      compositionSum += value;
      pm.expect(value, "Invalid quantity (must be positive [+0])").to.be.above(
        0
      );
    });
    if (compositionSum != 100) {
      pm.expect(pm.response.text(), "Invalid batch composition").to.not.include(
        `batch composition percentage sum should be equal to 100`
      );
    }

    // Final score
    pm.expect(
      parseFloat(finalScore),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
  });

  // Tests for production and batch creation
  pm.test("Batch & Production are added to state", function () {
    pm.expect(pm.response.text(), "Error adding batch").to.not.include(
      `failed to put batch to world state:`
    );
    pm.expect(pm.response.text(), "Error adding production").to.not.include(
      `failed to put production to world state:`
    );
    pm.expect(pm.response.text(), "Added production & batch").to.include(
      `production activity [${productionID}] & batch [${batchID}] were successfully added to the ledger`
    );
  });
}
