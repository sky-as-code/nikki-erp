const { testHttpResponse } = require('./common-test-response');
const { oneErrorSchemaNoField, oneErrorSchemaNoFieldWithVars, sameErrorSchema } = require('./common-utils');

const schema = oneErrorSchemaNoField('common.err_not_found', 'The desired data could not be found', 'business');

module.exports.testNotFound = function () {
  testHttpResponse(schema, 400);
};

module.exports.testFieldNotFound = function (field) {
  const schemaWithField = sameErrorSchema(field, 'common.err_not_found', 'The desired data could not be found', 'business');
  testHttpResponse([schemaWithField], 400);
};

module.exports.testValueNotFound = function (...values) {
  const varSchema = {
    type: 'object',
    required: ['values'],
    properties: {
      values: {
        type: 'array',
        minItems: values.length,
        maxItems: values.length,
        items: {
          type: 'string',
          enum: [...values],
        },
      },
    },
  };

  const schema = oneErrorSchemaNoFieldWithVars(
    'common.err_value_not_found', 'Value(s) could not be found: {.values}',
    varSchema, 'business',
  );

  testHttpResponse(schema, 400);
};
