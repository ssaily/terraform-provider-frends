package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildApiPolicySave_minimal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyEndpoints, diags := types.ListValue(types.ObjectType{AttrTypes: targetEndpointAttrTypes}, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty endpoints: %v", diags)
	}
	emptyIdentities, diags := types.ListValue(types.ObjectType{AttrTypes: identityAttrTypes}, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty identities: %v", diags)
	}
	emptyTags, diags := types.ListValue(types.StringType, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty tags: %v", diags)
	}

	model := ApiPolicyModel{
		Name:              types.StringValue("MyPolicy"),
		Description:       types.StringValue("A test policy"),
		AllowPublicAccess: types.BoolValue(false),
		ApiKeyName:        types.StringValue(""),
		ApiKeyLocation:    types.StringValue(""),
		Tags:              emptyTags,
		TargetEndpoints:   emptyEndpoints,
		Identities:        emptyIdentities,
	}

	result, diags := buildApiPolicySave(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if result.Name != "MyPolicy" {
		t.Fatalf("expected Name=MyPolicy, got %q", result.Name)
	}
	if result.AllowPublicAccess {
		t.Fatal("expected AllowPublicAccess=false")
	}
	if len(result.TargetEndpoints) != 0 {
		t.Fatalf("expected no endpoints, got %d", len(result.TargetEndpoints))
	}
	if len(result.Identities) != 0 {
		t.Fatalf("expected no identities, got %d", len(result.Identities))
	}
}

func TestBuildApiPolicySave_withEndpointsAndIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	endpointObj, diags := types.ObjectValue(targetEndpointAttrTypes, map[string]attr.Value{
		"url":    types.StringValue("https://my-company.frends.com/api/order"),
		"method": types.StringValue("POST"),
	})
	if diags.HasError() {
		t.Fatalf("building endpoint object: %v", diags)
	}
	endpoints, diags := types.ListValue(types.ObjectType{AttrTypes: targetEndpointAttrTypes}, []attr.Value{endpointObj})
	if diags.HasError() {
		t.Fatalf("building endpoints list: %v", diags)
	}

	ruleObj, diags := types.ObjectValue(identityRuleAttrTypes, map[string]attr.Value{
		"claim_type":  types.StringValue("aud"),
		"claim_value": types.StringValue("frends-api"),
		"match_type":  types.StringValue("Exact"),
	})
	if diags.HasError() {
		t.Fatalf("building rule object: %v", diags)
	}
	rulesList, diags := types.ListValue(types.ObjectType{AttrTypes: identityRuleAttrTypes}, []attr.Value{ruleObj})
	if diags.HasError() {
		t.Fatalf("building rules list: %v", diags)
	}

	identityObj, diags := types.ObjectValue(identityAttrTypes, map[string]attr.Value{
		"name":  types.StringValue("InternalServices"),
		"rules": rulesList,
	})
	if diags.HasError() {
		t.Fatalf("building identity object: %v", diags)
	}
	identities, diags := types.ListValue(types.ObjectType{AttrTypes: identityAttrTypes}, []attr.Value{identityObj})
	if diags.HasError() {
		t.Fatalf("building identities list: %v", diags)
	}

	emptyTags, diags := types.ListValue(types.StringType, []attr.Value{})
	if diags.HasError() {
		t.Fatalf("building empty tags: %v", diags)
	}

	model := ApiPolicyModel{
		Name:              types.StringValue("OrderPolicy"),
		Description:       types.StringValue(""),
		AllowPublicAccess: types.BoolValue(false),
		ApiKeyName:        types.StringValue("X-API-Key"),
		ApiKeyLocation:    types.StringValue("Header"),
		Tags:              emptyTags,
		TargetEndpoints:   endpoints,
		Identities:        identities,
	}

	result, diags := buildApiPolicySave(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if len(result.TargetEndpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(result.TargetEndpoints))
	}
	if result.TargetEndpoints[0].URL != "https://my-company.frends.com/api/order" {
		t.Fatalf("unexpected endpoint URL: %s", result.TargetEndpoints[0].URL)
	}
	if result.TargetEndpoints[0].Method != "POST" {
		t.Fatalf("expected method=POST, got %s", result.TargetEndpoints[0].Method)
	}

	if len(result.Identities) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(result.Identities))
	}
	if result.Identities[0].Name != "InternalServices" {
		t.Fatalf("unexpected identity name: %s", result.Identities[0].Name)
	}

	if len(result.Identities[0].Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result.Identities[0].Rules))
	}
	rule := result.Identities[0].Rules[0]
	if rule.ClaimType != "aud" || rule.ClaimValue != "frends-api" || rule.MatchType != "Exact" {
		t.Fatalf("unexpected rule: %+v", rule)
	}

	if result.ApiKeyName != "X-API-Key" || result.ApiKeyLocation != "Header" {
		t.Fatalf("unexpected API key fields: name=%s location=%s", result.ApiKeyName, result.ApiKeyLocation)
	}
}

func TestBuildApiPolicySave_publicAccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	emptyEndpoints, _ := types.ListValue(types.ObjectType{AttrTypes: targetEndpointAttrTypes}, []attr.Value{})
	emptyIdentities, _ := types.ListValue(types.ObjectType{AttrTypes: identityAttrTypes}, []attr.Value{})
	emptyTags, _ := types.ListValue(types.StringType, []attr.Value{})

	model := ApiPolicyModel{
		Name:              types.StringValue("PublicPolicy"),
		AllowPublicAccess: types.BoolValue(true),
		Tags:              emptyTags,
		TargetEndpoints:   emptyEndpoints,
		Identities:        emptyIdentities,
	}

	result, diags := buildApiPolicySave(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !result.AllowPublicAccess {
		t.Fatal("expected AllowPublicAccess=true")
	}
}
