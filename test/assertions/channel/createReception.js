// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json().message;

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
  var receptionID = req.args[0];
  productionUnitInternalID = req.args[1];
  activityDate = req.args[2];
  receivedBatchID = req.args[3];
  newBatchID = req.args[4];
  newBatchInternalID = req.args[5];
  isAccepted = req.args[6];
  transportScore = req.args[7];
  ses = req.args[8];
  distance = req.args[9];
  cost = req.args[10];

  // Tests for valid reception ID
  pm.test("Valid reception ID", function () {
    pm.expect(receptionID, "Empty reception ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `activity ID prefix must match its type (should be [rc-...])`
    );
    pm.expect(pm.response.text(), "Invalid activity ID").to.not.include(
      `incorrect activity prefix`
    );
    pm.expect(pm.response.text(), "Existing activity ID").to.not.include(
      `reception [${receptionID}] already exists`
    );
  });

  // Tests for valid timestamp
  pm.test("Got timestamp", function () {
    pm.expect(pm.response.text(), "Invalid timestamp").to.not.include(
      `could not get transaction timestamp:`
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
    pm.expect(pm.response.text(), "Invalid production unit ID").to.not.include(
      `must be different from batch's production unit ID`
    );
  });

  // Tests for valid issuer's ID
  pm.test("Got issuer's ID", function () {
    pm.expect(pm.response.text(), "Invalid issuer ID").to.not.include(
      `could not get issuer's client ID:`
    );
  });

  // Tests for valid received batch data
  pm.test("Received batch data is valid", function () {
    // Received batch ID
    pm.expect(receivedBatchID, "Empty received batch ID").to.not.be.empty;
    pm.expect(
      pm.response.text(),
      "Error reading received batch"
    ).to.not.include(`could not read batch from world state:`);
    pm.expect(pm.response.text(), "Invalid received batch ID").to.not.include(
      `batch [${receivedBatchID}] already exists`
    );
    pm.expect(
      pm.response.text(),
      "Invalid received batch transit state"
    ).to.not.include(`batch [${receivedBatchID}] is not in transit`);
  });

  // Tests for valid new batch data
  pm.test("New batch data is valid", function () {
    // New batch ID
    pm.expect(newBatchID, "Empty new batch ID").to.not.be.empty;
    pm.expect(pm.response.text(), "Invalid new batch ID").to.not.include(
      `batch [${newBatchID}] already exists`
    );
    pm.expect(pm.response.text(), "Invalid new batch ID prefix").to.not.include(
      `incorrect batch prefix. (should be [b-...])`
    );
    // Batch internal ID
    pm.expect(
      newBatchInternalID,
      "Empty new batch internal ID"
    ).to.not.be.empty;
    pm.expect(
      pm.response.text(),
      "Invalid new batch internal ID"
    ).to.not.include(`batch internal ID must not be empty`);
  });

  // Tests for valid scores
  pm.test("Valid activity scores", function () {
    pm.expect(
      parseFloat(transportScore),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
    pm.expect(
      parseFloat(ses),
      "Score out of bounds (should be between -10 & 10)"
    ).to.be.within(-10, 10);
  });

  // Tests for valid distance
  pm.test("Valid distance", function () {
    pm.expect(parseFloat(distance), "Invalid distance").to.be.at.least(0);
  });

  // Tests for valid cost
  pm.test("Valid cost", function () {
    pm.expect(parseFloat(cost), "Invalid cost").to.be.at.least(0);
  });

  // Tests for reception and batch creation
  pm.test("Batch & Reception are added to state", function () {
    pm.expect(pm.response.text(), "Error adding batch").to.not.include(
      `failed to put batch to world state:`
    );
    pm.expect(pm.response.text(), "Error adding reception").to.not.include(
      `failed to put reception to world state:`
    );
    pm.expect(pm.response.text(), "Added reception & batch").to.include(
      `reception activity [${receptionID}] & batch [${newBatchID}] were successfully added to the ledger. batch [${receivedBatchID}] was deleted successfully`
    );
  });
}
