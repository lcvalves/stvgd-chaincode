// Gets request and response bodies
var req = JSON.parse(pm.request.toJSON().body.raw);
res = pm.response.json();

// Sets admin bearer token to env variable
pm.environment.set("inovafil-bearer", res.token);

// Gets request credentials
var identity = req.id;
secret = req.secret;

// Tests if issuer has admin authorized credentials
pm.test("Has admin authorized credentials", function () {
  pm.expect(pm.response.text(), "Missing credentials").to.not.include(
    `is not set`
  );
  pm.expect(pm.response.text(), "Authentication failure").to.not.include(
    `Authentication failure`
  );
});
