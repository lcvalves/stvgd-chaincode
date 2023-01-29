// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json().message;

// Aux Registration variables
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
  var registrationID = req.args[0];
  productionUnitInternalID = req.args[1];
  batchID = req.args[2];
  batchType = req.args[3];
  batchInternalID = req.args[4];
  supplierID = req.args[5];
  unit = req.args[6];
  quantity = req.args[7];
  finalScore = req.args[8];
  batchComposition = req.args[9];
  batchCompositionMap = new Map(Object.entries(JSON.parse(batchComposition)));

  // Tests for valid registration ID
  pm.test("Valid registration ID", function () {
    pm.expect(registrationID, "Empty registration ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `activity ID prefix must match its type (should be [rg-...])`
    );
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `incorrect activity prefix`
    );
    pm.expect(pm.response.text(), "Existing activity ID").to.not.include(
      `registration [${registrationID}] already exists`
    );
  });

  // Tests for valid timestamp
  pm.test("Got timestamp", function () {
    pm.expect(pm.response.text(), "Invalid timestamp").to.not.include(
      `could not get transaction timestamp:`
    );
  });

  // Tests for valid batch type
  pm.test("Batch type is valid (FIBER, YARN, MESH, ...)", function () {
    pm.expect(batchType, "Empty batch type ID").to.not.be.empty;
    pm.expect(batchTypes, "Inserted batch type not defined").to.deep.include(
      batchType
    );
    pm.expect(pm.response.text(), "Invalid batch type").to.not.include(
      `could not validate batch type`
    );
  });

  // Tests for valid batch unit
  pm.test("Batch unit is valid (KG, L, M, ...)", function () {
    pm.expect(unit, "Empty unit").to.not.be.empty;
    pm.expect(units, "Inserted unit not defined").to.be.include(unit);
    pm.expect(pm.response.text(), "Invalid batch unit").to.not.include(
      `could not validate batch unit`
    );
  });

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

  // Tests for valid batch data
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

    // Supplier ID
    pm.expect(supplierID, "Empty supplier ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid supplier ID").to.not.include(
      `supplier ID must not be empty`
    );

    // Batch composition
    pm.expect(
      batchCompositionMap.size,
      "Invalid batch composition (must have at least 1 material)"
    ).to.be.at.least(1);
    batchCompositionMap.forEach((value, key) => {
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
    // Quantity
    pm.expect(parseFloat(quantity), "Invalid quantity").to.be.above(0);

    // Score
    pm.expect(
      parseFloat(finalScore),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
  });

  // Tests for registration and batch creation
  pm.test("Batch & Registration are added to state", function () {
    pm.expect(pm.response.text(), "Error adding batch").to.not.include(
      `failed to put batch to world state:`
    );
    pm.expect(pm.response.text(), "Error adding registration").to.not.include(
      `failed to put registration to world state:`
    );
    pm.expect(pm.response.text(), "Added registration & batch").to.include(
      `registration [${registrationID}] & batch [${batchID}] were successfully added to the ledger`
    );
  });
}
