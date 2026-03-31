const { testHttpResponse } = require('./common-test-response');

const schema = {
    type: "object",
    required: ["id", "created_at", "etag"],
    properties: {
        id: {
            type: "string",
            minLength: 1,
        },
        created_at: {
            type: "string",
            minLength: 20, //"2026-03-23T06:53:01Z"
        },
        etag: {
            type: "string",
            minLength: 1,
        },
    },
    additionalProperties: false
};

module.exports.testCreate = function (idVarName) {
  testHttpResponse(schema, 201);

  const { id, etag } = res.getBody();
  etag && bru.setVar('etag', etag);
  id && bru.setVar(idVarName, id);
};
