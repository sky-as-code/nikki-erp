const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

module.exports.testDuplicateValues = function (...fields) {  
  const schema = sameErrorSchema(fields, 'common.err_unique_constraint_violated', 'unique constraint violated {{.uniques}}', 'business');

  testHttpResponse(schema, 400);
};
