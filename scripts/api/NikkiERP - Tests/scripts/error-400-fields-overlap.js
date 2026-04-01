const { testHttpResponse } = require('./common-test-response');
const { oneErrorSchemaNoFieldWithVars } = require('./common-utils');

module.exports.testOverlappingFields = function (...fields) {
  const varSchema = {
    type: 'object',
    required: ['fields'],
    properties: {
      fields: {
        type: 'array',
        minItems: 2,
        maxItems: 2,
        items: {
          type: 'string',
          enum: ['add', 'remove'],
        },
      },
    },
  };

  const schema = oneErrorSchemaNoFieldWithVars("common.err_overlapped_fields", "These fields must not have overlapping values: {.fields}", varSchema, 'business');

  testHttpResponse(schema, 400);
};
