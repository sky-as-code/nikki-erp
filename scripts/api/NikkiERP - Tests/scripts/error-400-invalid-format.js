const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testInvalidFormat = function (...fields) {  
  const schema = sameErrorSchema(fields, "common.err_invalid_data_type", "invalid data type, must be a {{.typeName}}");

  testHttpResponse(schema, 400);
};
