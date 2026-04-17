package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandClaimsAndTags_allNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	model := PrivateApplicationModel{
		CustomTokenClaims: types.MapNull(types.StringType),
		Tags:              types.ListNull(types.StringType),
	}

	claims, tags, diags := expandClaimsAndTags(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(claims) != 0 {
		t.Fatalf("expected empty claims, got %v", claims)
	}
	if len(tags) != 0 {
		t.Fatalf("expected empty tags, got %v", tags)
	}
}

func TestExpandClaimsAndTags_withClaims(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	claimsMap, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"service": types.StringValue("order-service"),
		"env":     types.StringValue("production"),
	})
	if diags.HasError() {
		t.Fatalf("building claims map: %v", diags)
	}

	model := PrivateApplicationModel{
		CustomTokenClaims: claimsMap,
		Tags:              types.ListNull(types.StringType),
	}

	claims, tags, diags := expandClaimsAndTags(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if claims["service"] != "order-service" {
		t.Fatalf("expected claims[service]=order-service, got %v", claims["service"])
	}
	if claims["env"] != "production" {
		t.Fatalf("expected claims[env]=production, got %v", claims["env"])
	}
	if len(tags) != 0 {
		t.Fatalf("expected empty tags, got %v", tags)
	}
}

func TestExpandClaimsAndTags_withTags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tagList, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("production"),
		types.StringValue("critical"),
	})
	if diags.HasError() {
		t.Fatalf("building tag list: %v", diags)
	}

	model := PrivateApplicationModel{
		CustomTokenClaims: types.MapNull(types.StringType),
		Tags:              tagList,
	}

	claims, tags, diags := expandClaimsAndTags(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(claims) != 0 {
		t.Fatalf("expected empty claims, got %v", claims)
	}
	if len(tags) != 2 || tags[0] != "production" || tags[1] != "critical" {
		t.Fatalf("expected tags=[production, critical], got %v", tags)
	}
}

func TestExpandClaimsAndTags_withBoth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	claimsMap, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"aud": types.StringValue("frends-api"),
	})
	if diags.HasError() {
		t.Fatalf("building claims map: %v", diags)
	}
	tagList, diags := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("internal"),
	})
	if diags.HasError() {
		t.Fatalf("building tag list: %v", diags)
	}

	model := PrivateApplicationModel{
		CustomTokenClaims: claimsMap,
		Tags:              tagList,
	}

	claims, tags, diags := expandClaimsAndTags(ctx, model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if claims["aud"] != "frends-api" {
		t.Fatalf("expected claims[aud]=frends-api, got %v", claims["aud"])
	}
	if len(tags) != 1 || tags[0] != "internal" {
		t.Fatalf("expected tags=[internal], got %v", tags)
	}
}
