const tv4 = require('tv4');

module.exports.testHttpResponse = function (schema, httpStatus) {  
  test(`Status code is ${httpStatus}`, () => {
      expect(res.getStatus()).to.equal(httpStatus);
  });

  test("Response matches expected JSON schema", () => {
    // Uncomment for debugging. DO NOT COMMIT!
    // console.log({body: res.getBody()});
    const valid = tv4.validate(res.getBody(), schema);
    tv4.error && console.error(tv4.error);
    expect(tv4.error).not.to.exist;
    expect(valid).to.be.true;  
  });
};

module.exports.initTestData = function (prefix, modelSchema, nextRequest, loopCount=50) {
  test("Status code is 201", () => {
    expect(res.getStatus()).to.equal(201);
  });
  
  const schema = createSearchSchema(modelSchema);
  
  bru.setVar("search_schema", schema);
  bru.setVar("loop_count", loopCount);

  console.log(`${prefix}: Creating ${loopCount} test records`)
  bru.runner.setNextRequest(nextRequest);
};

module.exports.loopTestData = function (loopRequest, testCaseRequest) {
  test("Status code is 201", () => {
    expect(res.getStatus()).to.equal(201);
  });

  let i = bru.getVar("loop_count");

  if (i > 0) {
      bru.setVar("loop_count", --i);
      console.log("To loop: ", i)
      bru.runner.setNextRequest(loopRequest);
  } else {
      console.log("To test case")
      bru.deleteVar("loop_count")
      bru.runner.setNextRequest(testCaseRequest);
  }
};

function createSearchSchema(modelSchema) {
  return {
    type: "object",
    required: ["items", "total", "page", "size"],
    properties: {
      items: {
        type: "array",
        items: {
          type: "object",
          properties: modelSchema,
        },
      },
      total: {
        type: "integer",
        minimum: 0,
      },
      page: {
        type: "integer",
        minimum: 0,
      },
      size: {
        type: "integer",
        minimum: 50,
      },
    },
  };
};
