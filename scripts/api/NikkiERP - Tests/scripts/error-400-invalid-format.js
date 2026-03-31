const { testHttpResponse } = require('./common-test-response');

module.exports.testInvalidFormat = function (...fields) {  
  const schema = {
      type: "array",
      minItems: fields.length,
      maxItems: fields.length,
      items: {
          type: "object",
          required: ["field", "message", "type"],
          properties: {
              field: {
                  type: "string",
                  enum: [...fields],
              },
              key: {
                  type: "string",
                  enum: ["common.err_invalid_data_type"],
              },
              message: {
                  type: "string",
                  enum: ["invalid data type"],
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
