const { testHttpResponse } = require('./common-test-response');

const schema = {
    type: "array",
    minItems: 1,
    maxItems: 1,
    items: {
        type: "object",
        required: ["key", "message", "type"],
        properties: {
            key: {
                type: "string",
                enum: ["common.err_not_found"],
            },
            message: {
                type: "string",
                enum: ["The desired data could not be found"],
            },
            type: {
                type: "string",
                enum: ["business"],
            },
        },
        additionalProperties: false,
    },
    additionalItems: false
};

module.exports.testNotFound = function () {
  testHttpResponse(schema, 400);
};
