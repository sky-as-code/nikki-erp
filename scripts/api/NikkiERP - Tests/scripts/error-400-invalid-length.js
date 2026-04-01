const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testInvalidStringLength = function (...fields) {  
  const schema = sameErrorSchema(fields, "common.err_invalid_string_length", "string length must be between {{.min}} and {{.max}}");

  testHttpResponse(schema, 400);
};

module.exports.testInvalidArrayLength = function (...fields) {  
  const schema = sameErrorSchema(fields, "common.err_invalid_array_length", "array length must be between {{.min}} and {{.max}}");

  testHttpResponse(schema, 400);
};
