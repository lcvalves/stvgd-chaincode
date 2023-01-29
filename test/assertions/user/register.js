// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json().message;

// Gets request credentials
var identity = req.id;
secret = req.secret;

if (res == "User with provided token is not enrolled") {
  // Tests if issuer is enrolled
  pm.test("User is enrolled ", function () {
    pm.expect(pm.response.text(), "Missing authentication").to.not.include(
      `User with provided token is not enrolled`
    );
  });
} else {
  // Tests if issuer is enrolled
  pm.test("User is enrolled ", function () {
    pm.expect(pm.response.text(), "Missing authentication").to.not.include(
      `User with provided token is not 	enrolled`
    );
  });

  // Tests if issuer is admin
  pm.test("Has permission to register", function () {
    pm.expect(pm.response.text(), "Missing admin authorization").to.not.include(
      `Missing authorization header`
    );
  });

  // Tests if new client ID already exists
  pm.test("Valid identity name", function () {
    pm.expect(pm.response.text(), "Invalid identity name").to.not.include(
      `Identity '${identity}' is already registered`
    );
    pm.expect(pm.response.text(), "Invalid identity name").to.not.include(
      `Missing required argument`
    );
  });

  // Tests if new identity is added to CA
  pm.test("Added identity to CA", function () {
    pm.expect(pm.response.to.have.status(201));
    pm.expect(pm.response.text(), "Added identity to CA").to.include(`ok`);
  });
}
