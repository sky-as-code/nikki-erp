const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testNonExistFields = function (...fields) {
  const schema = sameErrorSchema(fields, "common.err_unknown_schema_field", "field is not defined on this schema");

  testHttpResponse(schema, 400);
};
