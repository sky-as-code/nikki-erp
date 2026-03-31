const { testHttpResponse } = require('./common-test-response');

const schema = {
    type: "array",
    minItems: 1,
    maxItems: 1,
    items: {
        type: "object",
        required: ["field", "key", "message", "type"],
        properties: {
            field: {
                type: "string",
                enum: ["etag"],
            },
            key: {
                type: "string",
                enum: ["common.err_etag_mismatched"],
            },
            message: {
                type: "string",
                enum: ["This data has been modified by another process"],
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

module.exports.testUnmatchedEtag = function () {
  testHttpResponse(schema, 400);
};
