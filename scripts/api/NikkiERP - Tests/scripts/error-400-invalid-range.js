const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testInvalidNumberRange = function (...fields) {  
  const schema = sameErrorSchema(fields, "common.err_invalid_number_range", "value must be between {{.min}} and {{.max}}");

  testHttpResponse(schema, 400);
};
