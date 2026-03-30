const { testHttpResponse } = require('./common-test-response');

module.exports.testDelete = function (count) {  
  const schema = {
      type: "object",
      required: ["affected_count", "affected_at"],
      properties: {
          affected_count: {
              type: "integer",
              enum: [count],
          },
          affected_at: {
              type: "string",
              minLength: 20, //"2026-03-23T06:53:01Z"
          },
      },
      additionalProperties: false
  };

  testHttpResponse(schema, 200);
};
