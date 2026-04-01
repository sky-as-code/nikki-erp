const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testMissingFields = function (...fields) {
  const schema = sameErrorSchema(fields, "common.err_missing_required_field", "field is required");

  testHttpResponse(schema, 400);
};
