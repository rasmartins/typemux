"""
Validation tests for generated IDL schemas.

This test suite validates that the generated Protobuf, GraphQL, and OpenAPI
schemas are valid according to their respective specifications.
"""

import os
import subprocess
import pytest
import yaml
from pathlib import Path
from graphql import build_schema
from openapi_spec_validator import validate_spec


# Get the project root directory (parent of validate/)
PROJECT_ROOT = Path(__file__).parent.parent.parent
EXAMPLES_DIR = PROJECT_ROOT / "examples"

# Example directories to test
EXAMPLE_DIRS = [
    "basic/generated",
    "custom-field-numbers/generated",
    "status-codes/generated",
    "unions/generated",
    "imports/generated",
    "namespaces/generated",
    "cross-namespace/generated",
    "name-annotations/generated",
]


class TestProtobufValidation:
    """Test Protobuf schema validation using protoc."""

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_protobuf_schema_exists(self, example_dir):
        """Verify that the Protobuf schema file exists."""
        proto_file = EXAMPLES_DIR / example_dir / "schema.proto"
        assert proto_file.exists(), f"Protobuf schema not found at {proto_file}"

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_protobuf_schema_valid(self, example_dir):
        """Validate Protobuf schema using protoc."""
        proto_file = EXAMPLES_DIR / example_dir / "schema.proto"

        # Use protoc to validate the schema - use relative path from the generated folder
        result = subprocess.run(
            ["protoc", "--proto_path=.", "--descriptor_set_out=/dev/null", "schema.proto"],
            cwd=proto_file.parent,
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, f"Protobuf validation failed for {example_dir}:\n{result.stderr}"


class TestGraphQLValidation:
    """Test GraphQL schema validation."""

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_graphql_schema_exists(self, example_dir):
        """Verify that the GraphQL schema file exists."""
        graphql_file = EXAMPLES_DIR / example_dir / "schema.graphql"
        assert graphql_file.exists(), f"GraphQL schema not found at {graphql_file}"

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_graphql_schema_valid(self, example_dir):
        """Validate GraphQL schema using graphql-core."""
        graphql_file = EXAMPLES_DIR / example_dir / "schema.graphql"

        with open(graphql_file, 'r') as f:
            schema_content = f.read()

        try:
            schema = build_schema(schema_content)
            assert schema is not None, f"Failed to build GraphQL schema for {example_dir}"
        except Exception as e:
            pytest.fail(f"GraphQL schema validation failed for {example_dir}: {str(e)}")

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_graphql_has_query_type(self, example_dir):
        """Verify that the GraphQL schema has a Query type."""
        graphql_file = EXAMPLES_DIR / example_dir / "schema.graphql"

        with open(graphql_file, 'r') as f:
            schema_content = f.read()

        schema = build_schema(schema_content)
        query_type = schema.query_type

        assert query_type is not None, f"GraphQL schema for {example_dir} should have a Query type"

    @pytest.mark.parametrize("example_dir", ["basic/generated", "status-codes/generated"])
    def test_graphql_has_mutation_type(self, example_dir):
        """Verify that GraphQL schemas with services have a Mutation type."""
        graphql_file = EXAMPLES_DIR / example_dir / "schema.graphql"

        with open(graphql_file, 'r') as f:
            schema_content = f.read()

        schema = build_schema(schema_content)
        mutation_type = schema.mutation_type

        assert mutation_type is not None, f"GraphQL schema for {example_dir} should have a Mutation type"


class TestOpenAPIValidation:
    """Test OpenAPI schema validation."""

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_schema_exists(self, example_dir):
        """Verify that the OpenAPI schema file exists."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"
        assert openapi_file.exists(), f"OpenAPI schema not found at {openapi_file}"

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_yaml_parseable(self, example_dir):
        """Verify that the OpenAPI YAML is parseable."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"

        with open(openapi_file, 'r') as f:
            try:
                spec = yaml.safe_load(f)
                assert spec is not None, f"Failed to parse OpenAPI YAML for {example_dir}"
            except yaml.YAMLError as e:
                pytest.fail(f"OpenAPI YAML parsing failed for {example_dir}: {str(e)}")

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_schema_valid(self, example_dir):
        """Validate OpenAPI schema using openapi-spec-validator."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"

        with open(openapi_file, 'r') as f:
            spec = yaml.safe_load(f)

        try:
            validate_spec(spec)
        except Exception as e:
            pytest.fail(f"OpenAPI validation failed for {example_dir}: {str(e)}")

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_has_paths(self, example_dir):
        """Verify that the OpenAPI schema has paths defined."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"

        with open(openapi_file, 'r') as f:
            spec = yaml.safe_load(f)

        assert 'paths' in spec, f"OpenAPI schema for {example_dir} should have paths"
        assert len(spec['paths']) > 0, f"OpenAPI schema for {example_dir} should have at least one path"

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_has_components(self, example_dir):
        """Verify that the OpenAPI schema has components defined."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"

        with open(openapi_file, 'r') as f:
            spec = yaml.safe_load(f)

        assert 'components' in spec, f"OpenAPI schema for {example_dir} should have components"
        assert 'schemas' in spec['components'], f"OpenAPI schema for {example_dir} should have schemas in components"

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_openapi_version(self, example_dir):
        """Verify that the OpenAPI schema specifies version 3.0.0."""
        openapi_file = EXAMPLES_DIR / example_dir / "openapi.yaml"

        with open(openapi_file, 'r') as f:
            spec = yaml.safe_load(f)

        assert 'openapi' in spec, f"OpenAPI schema for {example_dir} should specify version"
        assert spec['openapi'].startswith('3.0'), f"OpenAPI schema for {example_dir} should be version 3.0.x"


class TestCrossSchemaConsistency:
    """Test consistency across different schema formats."""

    @pytest.mark.parametrize("example_dir", EXAMPLE_DIRS)
    def test_all_schemas_exist(self, example_dir):
        """Verify that all three schema types exist for each example."""
        base_path = EXAMPLES_DIR / example_dir

        proto_file = base_path / "schema.proto"
        graphql_file = base_path / "schema.graphql"
        openapi_file = base_path / "openapi.yaml"

        assert proto_file.exists(), f"Protobuf schema missing for {example_dir}"
        assert graphql_file.exists(), f"GraphQL schema missing for {example_dir}"
        assert openapi_file.exists(), f"OpenAPI schema missing for {example_dir}"
