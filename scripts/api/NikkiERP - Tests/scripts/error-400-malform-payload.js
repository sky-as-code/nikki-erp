const { testHttpResponse } = require('./common-test-response');
const { oneErrorSchemaNoField } = require('./common-utils');

const schema = oneErrorSchemaNoField("common.err_malformed_request", "malformed request");

module.exports.testMalformPayload = function () {  
  testHttpResponse(schema, 400);
};
