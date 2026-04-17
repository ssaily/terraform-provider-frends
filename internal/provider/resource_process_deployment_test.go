package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandProcessVersions_nullList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	list := types.ListNull(types.ObjectType{AttrTypes: processVersionAttrTypes})

	result, diags := expandProcessVersions(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d elements", len(result))
	}
}

func TestExpandProcessVersions_singleProcessNoVariables(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyVars, diags := types.ListValue(
		types.ObjectType{AttrTypes: processVariableAttrTypes}, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty vars list: %v", diags)
	}

	procObj, diags := types.ObjectValue(processVersionAttrTypes, map[string]attr.Value{
		"process_guid":      types.StringValue("aaaaaaaa-0000-0000-0000-000000000001"),
		"version":           types.Int64Value(5),
		"process_variables": emptyVars,
	})
	if diags.HasError() {
		t.Fatalf("building process object: %v", diags)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: processVersionAttrTypes}, []attr.Value{procObj})
	if diags.HasError() {
		t.Fatalf("building process list: %v", diags)
	}

	result, diags := expandProcessVersions(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 process, got %d", len(result))
	}
	if result[0].ProcessGUID != "aaaaaaaa-0000-0000-0000-000000000001" {
		t.Fatalf("unexpected GUID: %s", result[0].ProcessGUID)
	}
	if result[0].Version != 5 {
		t.Fatalf("expected Version=5, got %d", result[0].Version)
	}
	if len(result[0].ProcessVariables) != 0 {
		t.Fatalf("expected no variables, got %d", len(result[0].ProcessVariables))
	}
}

func TestExpandProcessVersions_withVariables(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	varObj, diags := types.ObjectValue(processVariableAttrTypes, map[string]attr.Value{
		"name":        types.StringValue("ApiBaseUrl"),
		"value":       types.StringValue("https://api.example.com"),
		"is_secret":   types.BoolValue(false),
		"description": types.StringValue("Base URL for the API"),
	})
	if diags.HasError() {
		t.Fatalf("building variable object: %v", diags)
	}

	varsList, diags := types.ListValue(types.ObjectType{AttrTypes: processVariableAttrTypes}, []attr.Value{varObj})
	if diags.HasError() {
		t.Fatalf("building variables list: %v", diags)
	}

	procObj, diags := types.ObjectValue(processVersionAttrTypes, map[string]attr.Value{
		"process_guid":      types.StringValue("bbbbbbbb-0000-0000-0000-000000000002"),
		"version":           types.Int64Value(1),
		"process_variables": varsList,
	})
	if diags.HasError() {
		t.Fatalf("building process object: %v", diags)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: processVersionAttrTypes}, []attr.Value{procObj})
	if diags.HasError() {
		t.Fatalf("building process list: %v", diags)
	}

	result, diags := expandProcessVersions(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 process, got %d", len(result))
	}
	if len(result[0].ProcessVariables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(result[0].ProcessVariables))
	}
	v := result[0].ProcessVariables[0]
	if v.Name != "ApiBaseUrl" {
		t.Fatalf("expected Name=ApiBaseUrl, got %s", v.Name)
	}
	if v.Value != "https://api.example.com" {
		t.Fatalf("expected correct value, got %s", v.Value)
	}
	if v.IsSecret {
		t.Fatal("expected IsSecret=false")
	}
}

func TestExpandProcessVersions_secretVariable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	varObj, diags := types.ObjectValue(processVariableAttrTypes, map[string]attr.Value{
		"name":        types.StringValue("DbPassword"),
		"value":       types.StringValue("s3cr3t"),
		"is_secret":   types.BoolValue(true),
		"description": types.StringValue(""),
	})
	if diags.HasError() {
		t.Fatalf("building variable object: %v", diags)
	}

	varsList, diags := types.ListValue(types.ObjectType{AttrTypes: processVariableAttrTypes}, []attr.Value{varObj})
	if diags.HasError() {
		t.Fatalf("building variables list: %v", diags)
	}

	emptyVarsList, diags := types.ListValue(types.ObjectType{AttrTypes: processVariableAttrTypes}, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty variables list: %v", diags)
	}
	_ = emptyVarsList

	procObj, diags := types.ObjectValue(processVersionAttrTypes, map[string]attr.Value{
		"process_guid":      types.StringValue("cccccccc-0000-0000-0000-000000000003"),
		"version":           types.Int64Value(2),
		"process_variables": varsList,
	})
	if diags.HasError() {
		t.Fatalf("building process object: %v", diags)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: processVersionAttrTypes}, []attr.Value{procObj})
	if diags.HasError() {
		t.Fatalf("building process list: %v", diags)
	}

	result, diags := expandProcessVersions(ctx, list)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !result[0].ProcessVariables[0].IsSecret {
		t.Fatal("expected IsSecret=true for secret variable")
	}
}
