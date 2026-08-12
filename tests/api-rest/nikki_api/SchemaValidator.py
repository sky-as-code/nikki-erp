"""Draft-4 JSON-schema validation and schema factories for API REST tests.

This is the only custom Python library in the test framework (see rule 3 in
"docs/06. API testing.md"). The schema factories are verbatim ports of the Bruno
collection's "scripts/common-utils.js" and the per-module schemas embedded in
"scripts/*.js"; response validation mirrors "common-test-response.js" (tv4).
"""
import json
from pathlib import Path

from jsonschema import Draft4Validator, FormatChecker


class SchemaValidator:
    ROBOT_LIBRARY_SCOPE = "GLOBAL"

    def validate_json_schema(self, instance, schema):
        """Validate ``instance`` against a draft-4 ``schema`` (dict or path to a .json file)."""
        if isinstance(schema, (str, Path)):
            schema = json.loads(Path(schema).read_text(encoding="utf-8"))
        validator = Draft4Validator(schema, format_checker=FormatChecker())
        errors = sorted(validator.iter_errors(instance), key=lambda err: list(err.path))
        if errors:
            details = "\n".join(
                "- {}: {}".format("/".join(str(p) for p in err.path) or "(root)", err.message)
                for err in errors
            )
            raise AssertionError(
                "JSON schema validation failed:\n{}\nActual payload:\n{}".format(
                    details, json.dumps(instance, indent=2, ensure_ascii=False)
                )
            )

    def same_error_schema(self, fields, key, message, type="validation"):
        """Error array where every item targets one of ``fields`` with the same key/message."""
        fields = list(fields)
        return {
            "type": "array",
            "minItems": len(fields),
            "maxItems": len(fields),
            "items": {
                "type": "object",
                "required": ["field", "key", "message", "type"],
                "properties": {
                    "field": {"type": "string", "enum": fields},
                    "key": {"type": "string", "enum": [key]},
                    "message": {"type": "string", "enum": [message]},
                    "type": {"type": "string", "enum": [type]},
                    "vars": {"type": "object"},
                },
                "additionalProperties": False,
            },
            "additionalItems": False,
        }

    def one_error_schema_no_field(self, key, message, type="validation", vars_schema=None):
        """Single-item error array without a ``field`` property (optionally pinning ``vars``)."""
        return {
            "type": "array",
            "minItems": 1,
            "maxItems": 1,
            "items": {
                "type": "object",
                "required": ["key", "message", "type"],
                "properties": {
                    "key": {"type": "string", "enum": [key]},
                    "message": {"type": "string", "enum": [message]},
                    "type": {"type": "string", "enum": [type]},
                    "vars": vars_schema or {"type": "object"},
                },
                "additionalProperties": False,
            },
            "additionalItems": False,
        }

    def mutate_success_schema(self, count):
        """Update/archive/delete response: {affected_count==count, affected_at, etag?}."""
        return {
            "type": "object",
            "required": ["affected_count", "affected_at"],
            "properties": {
                "affected_count": {"type": "integer", "enum": [int(count)]},
                "affected_at": {"type": "string", "minLength": 20},
                "etag": {"type": "string", "minLength": 1},
            },
            "additionalProperties": False,
        }

    def delete_success_schema(self, count):
        schema = self.mutate_success_schema(count)
        del schema["properties"]["etag"]
        return schema

    def exists_success_schema(self, existing, not_existing):
        return {
            "type": "object",
            "required": ["existing", "not_existing"],
            "properties": {
                "existing": {
                    "type": "array",
                    "minItems": int(existing),
                    "maxItems": int(existing),
                    "items": {"type": "string"},
                },
                "not_existing": {
                    "type": "array",
                    "minItems": int(not_existing),
                    "maxItems": int(not_existing),
                },
            },
            "additionalProperties": False,
        }

    def search_success_schema(self, model_schema):
        """Search envelope around a model schema (dict or path to a .json file).

        Items are validated against the model's ``properties`` only (no ``required``,
        no ``additionalProperties``) because search supports column projections.
        """
        if isinstance(model_schema, (str, Path)):
            model_schema = json.loads(Path(model_schema).read_text(encoding="utf-8"))
        item_schema = {"type": "object", "properties": model_schema.get("properties", {})}
        return {
            "type": "object",
            "required": ["items", "total", "page", "size"],
            "properties": {
                "items": {"type": "array", "items": item_schema},
                "total": {"type": "integer", "minimum": 0},
                "page": {"type": "integer", "minimum": 0},
                "size": {"type": "integer", "minimum": 1},
            },
        }
