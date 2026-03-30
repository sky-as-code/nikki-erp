const { testHttpResponse } = require('./common-test-response');

module.exports.testInvalidLength = function (...fields) {  
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
                enum: ["common.err_invalid_length"],
              },
              message: {
                  type: "string",
                  enum: ["must have length between {{.min}} and {{.max}}", "must be a valid email address"],
              },
              type: {
                  type: "string",
                  enum: ["validation"],
              },
              vars: {
                  type: "object",
                  properties: {
                    max: {
                      type: "integer",
                    },
                    min: {
                      type: "integer",
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
