const { testHttpResponse } = require('./common-test-response');
const { oneErrorSchemaNoFieldWithVars } = require('./common-utils');

module.exports.testDuplicateValues = function (...uniques) {
  const varsSchema = {
      type: "object",
      properties: {
        uniques: {
          type: "array",
        },
      },
    };
  const schema = oneErrorSchemaNoFieldWithVars('common.err_unique_constraint_violated', 'unique constraint violated {{.uniques}}', varsSchema, 'business');

  testHttpResponse(schema, 400);
  
  test("Correct unique fields", () => {
    const body = res.getBody();
    if (!Array.isArray(body) && Boolean(body[0])) {
      expect.fail('Response body is not duplicate error');
      return;
    }
    const { vars : { uniques: actual }} = body[0];
    expect(actual).to.have.deep.equal(uniques);
  });
};
