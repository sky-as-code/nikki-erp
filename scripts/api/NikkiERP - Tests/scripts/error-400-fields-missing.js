const { testHttpResponse } = require('./common-test-response');

module.exports.testMissingFields = function (...fields) {
  
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
                  enum: ["common.err_missing_required_field"],
              },
              message: {
                  type: "string",
                  enum: ["field is required"],
              },
              type: {
                  type: "string",
                  enum: ["validation"],
              },
          },
          additionalProperties: false,
      },
      additionalItems: false
  };

  testHttpResponse(schema, 400);
};
