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

function initTestData(prefix, modelSchema, nextRequest, loopCount=50) {
  test("Status code is 201", () => {
    expect(res.getStatus()).to.equal(201);
  });
  
  const schema = createSearchSchema(modelSchema);
  
  bru.setVar("search_schema", schema);
  bru.setVar("loop_count", loopCount-1);

  // console.log(`${prefix}: Creating ${loopCount} test records`);
  bru.runner.setNextRequest(nextRequest);
}

module.exports.initTestData = initTestData;

module.exports.initTestDataWithIds = function(prefix, modelSchema, nextRequest, idsVarName, loopCount=50) {
  initTestData(prefix, modelSchema, nextRequest, loopCount);

  const payload = res.getBody();
  const { id } = payload;
  bru.setVar(idsVarName, [id]);
};

function loopTestData(loopRequest, testCaseRequest) {
  test("Status code is 201", () => {
    expect(res.getStatus()).to.equal(201);
  });

  let i = bru.getVar("loop_count");

  if (i > 1) {
      bru.setVar("loop_count", --i);
      // console.log("To loop: ", i);
      bru.runner.setNextRequest(loopRequest);
  } else {
      // console.log("To test case");
      bru.deleteVar("loop_count");
      bru.runner.setNextRequest(testCaseRequest);
  }
}

module.exports.loopTestData = loopTestData;

module.exports.loopTestDataWithIds = function (loopRequest, testCaseRequest, idsVarName) {
  loopTestData(loopRequest, testCaseRequest);

  const idArr = bru.getVar(idsVarName);
  if (!Array.isArray(idArr)) {
    throw new Error(`loopTestDataWithIds: ${idsVarName} must be an array`);  
  }
  const { id } = res.getBody();
  idArr.push(id);
  bru.setVar(idsVarName, idArr);
};

module.exports.getTestDataIds = function (idsVarName) {
  const ids = bru.getVar(idsVarName) || []

  test("Test data was prepared", function () {
      expect(Boolean(Array.isArray(ids) && ids.length)).to.equal(true, "Must create test data before running this request");
  });
  
  return ids;
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
        minimum: 1,
      },
    },
  };
}
