const { testHttpResponse } = require('./common-test-response');

module.exports.testDuplicateValues = function (...fields) {  
  const schema = {
      type: "array",
      minItems: fields.length,
      maxItems: fields.length,
      items: {
          type: "object",
          required: ["field", "key", "message", "type"],
          properties: {
              field: {
                  type: "string",
                  enum: [...fields],
              },
              key: {
                  type: "string",
                  enum: ["common.err_unique_constraint_violated"],
              },
              message: {
                  type: "string",
                  enum: ["unique constraint violated {{.uniques}}"],
              },
              type: {
                  type: "string",
                  enum: ["business"],
              },
              vars: {
                  type: "object",
                  properties: {
                    uniques: {
                      type: "array",
                    },
                  },
              },
          },
          additionalProperties: false,
      },
      additionalItems: false
  };

  testHttpResponse(schema, 400);
};
