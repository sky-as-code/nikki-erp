const { testHttpResponse } = require('./common-test-response');

module.exports.testUpdate = function (count) {  
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
          etag: {
              type: "string",
              minLength: 1,
          },
      },
      additionalProperties: false
  };

  testHttpResponse(schema, 200);

  const { etag } = res.getBody();
  const previousEtag = bru.getVar("etag");

  test("Response 'etag' is different from previous", () => {
      expect(etag).to.not.eql(previousEtag);
  });

  etag && bru.setVar('etag', etag);
};
