module.exports.sameErrorSchema = function (fields, key, message, type='validation') {
  return {
    type: "array",
    minItems: fields.length,
    maxItems: fields.length,
    items: {
      type: "object",
      required: ["field", "key", "message", "type"],
      properties: {
        field: {
          type: "string",
          enum: [...fields],
        },
        key: {
          type: "string",
          enum: [key],
        },
        message: {
          type: "string",
          enum: [message],
        },
        type: {
          type: "string",
          enum: [type],
        },
        vars: {
          type: "object",
        },
        
      },
      additionalProperties: false,
    },
    additionalItems: false,
  };
};

module.exports.oneErrorSchemaNoField = function (key, message, type='validation') {
  return {
    type: "array",
    minItems: 1,
    maxItems: 1,
    items: {
      type: "object",
      required: ["key", "message", "type"],
      properties: {
        key: {
          type: "string",
          enum: [key],
        },
        message: {
          type: "string",
          enum: [message],
        },
        type: {
          type: "string",
          enum: [type],
        },
        vars: {
          type: "object",
        },
      },
      additionalProperties: false,
    },
    additionalItems: false,
  };
};


module.exports.oneErrorSchemaNoFieldWithVars = function (key, message, varSchema, type='validation') {
  return {
    type: "array",
    minItems: 1,
    maxItems: 1,
    items: {
      type: "object",
      required: ["key", "message", "type"],
      properties: {
        key: {
          type: "string",
          enum: [key],
        },
        message: {
          type: "string",
          enum: [message],
        },
        type: {
          type: "string",
          enum: [type],
        },
        vars: varSchema,
      },
      additionalProperties: false,
    },
    additionalItems: false,
  };
};