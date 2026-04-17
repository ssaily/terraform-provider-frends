package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandEnvVarValues_nullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	list := types.ListNull(types.ObjectType{AttrTypes: envVarValueAttrTypes})

	result, diags := expandEnvVarValues(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result for null list, got %d elements", len(result))
	}
}

func TestExpandEnvVarValues_unknownList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	list := types.ListUnknown(types.ObjectType{AttrTypes: envVarValueAttrTypes})

	result, diags := expandEnvVarValues(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result for unknown list, got %d elements", len(result))
	}
}

func TestExpandEnvVarValues_singleValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	obj, diags := types.ObjectValue(envVarValueAttrTypes, map[string]attr.Value{
		"environment_id": types.Int64Value(42),
		"value":          types.StringValue("hello"),
	})
	if diags.HasError() {
		t.Fatalf("building test object: %v", diags)
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: envVarValueAttrTypes}, []attr.Value{obj})
	if diags.HasError() {
		t.Fatalf("building test list: %v", diags)
	}

	result, diags := expandEnvVarValues(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	if result[0].EnvironmentID != 42 {
		t.Fatalf("expected EnvironmentID=42, got %d", result[0].EnvironmentID)
	}
	if result[0].Value != "hello" {
		t.Fatalf("expected Value=hello, got %v", result[0].Value)
	}
}

func TestExpandEnvVarValues_multipleValues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		envID int64
		value string
	}{
		{1, "dev-value"},
		{2, "staging-value"},
		{3, "prod-value"},
	}

	objs := make([]attr.Value, len(cases))
	for i, c := range cases {
		obj, diags := types.ObjectValue(envVarValueAttrTypes, map[string]attr.Value{
			"environment_id": types.Int64Value(c.envID),
			"value":          types.StringValue(c.value),
		})
		if diags.HasError() {
			t.Fatalf("building object %d: %v", i, diags)
		}
		objs[i] = obj
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: envVarValueAttrTypes}, objs)
	if diags.HasError() {
		t.Fatalf("building list: %v", diags)
	}

	result, diags := expandEnvVarValues(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	for i, c := range cases {
		if result[i].EnvironmentID != c.envID {
			t.Errorf("[%d] expected EnvironmentID=%d, got %d", i, c.envID, result[i].EnvironmentID)
		}
		if result[i].Value != c.value {
			t.Errorf("[%d] expected Value=%q, got %v", i, c.value, result[i].Value)
		}
	}
}
