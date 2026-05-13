const { testHttpResponse } = require('./common-test-response');
const { sameErrorSchema } = require('./common-utils');

const schema = sameErrorSchema(['etag'], 'common.err_etag_mismatched', 'This data has been modified by someone else', 'business');

module.exports.testUnmatchedEtag = function () {
  testHttpResponse(schema, 400);
};
