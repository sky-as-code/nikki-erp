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
                enum: ["common.err_malformed_request"],
            },
            message: {
                type: "string",
                enum: ["malformed request"],
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

module.exports.testMalformPayload = function () {  
  testHttpResponse(schema, 400);
};
