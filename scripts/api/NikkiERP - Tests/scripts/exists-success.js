const { testHttpResponse } = require('./common-test-response');

module.exports.testExists = function (existingCount, notExistingCount) {
  
const schema = {
  type: "object",
  required: ["existing", "not_existing"],
  properties: {
    existing: {
      type: "array",
      minItems: existingCount,
      maxItems: existingCount,
      items: {
        type: "string",
      },
    },
    not_existing: {
      type: "array",
      minItems: notExistingCount,
      maxItems: notExistingCount,
    },
  },
  additionalProperties: false
};

testHttpResponse(schema, 200);
};
