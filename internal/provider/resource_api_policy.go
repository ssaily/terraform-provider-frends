package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ApiPolicyResource{}
var _ resource.ResourceWithImportState = &ApiPolicyResource{}

func NewApiPolicyResource() resource.Resource {
	return &ApiPolicyResource{}
}

type ApiPolicyResource struct {
	client *client.Client
}

type ApiPolicyModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AllowPublicAccess types.Bool   `tfsdk:"allow_public_access"`
	ApiKeyName        types.String `tfsdk:"api_key_name"`
	ApiKeyLocation    types.String `tfsdk:"api_key_location"`
	Tags              types.List   `tfsdk:"tags"`
	TargetEndpoints   types.List   `tfsdk:"target_endpoints"`
	Identities        types.List   `tfsdk:"identities"`
}

type apiPolicyTargetEndpointTF struct {
	URL    types.String `tfsdk:"url"`
	Method types.String `tfsdk:"method"`
}

type apiPolicyIdentityTF struct {
	Name  types.String `tfsdk:"name"`
	Rules types.List   `tfsdk:"rules"`
}

type apiPolicyIdentityRuleTF struct {
	ClaimType  types.String `tfsdk:"claim_type"`
	ClaimValue types.String `tfsdk:"claim_value"`
	MatchType  types.String `tfsdk:"match_type"`
}

var targetEndpointAttrTypes = map[string]attr.Type{
	"url":    types.StringType,
	"method": types.StringType,
}

var identityRuleAttrTypes = map[string]attr.Type{
	"claim_type":  types.StringType,
	"claim_value": types.StringType,
	"match_type":  types.StringType,
}

var identityAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"rules": types.ListType{ElemType: types.ObjectType{AttrTypes: identityRuleAttrTypes}},
}

func (r *ApiPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_policy"
}

func (r *ApiPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends API Policy that controls access to API endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the API policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the API policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the API policy.",
				Optional:    true,
			},
			"allow_public_access": schema.BoolAttribute{
				Description: "Whether public access is allowed without authentication.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"api_key_name": schema.StringAttribute{
				Description: "Name of the API key parameter.",
				Optional:    true,
			},
			"api_key_location": schema.StringAttribute{
				Description: "Location of the API key: Header or Query.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags for the policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"target_endpoints": schema.ListNestedBlock{
				Description: "Target API endpoints this policy applies to.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Description: "Target endpoint URL.",
							Required:    true,
						},
						"method": schema.StringAttribute{
							Description: "HTTP method (GET, POST, etc.).",
							Optional:    true,
						},
					},
				},
			},
			"identities": schema.ListNestedBlock{
				Description: "Identities allowed to access the endpoints.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Identity name.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"rules": schema.ListNestedBlock{
							Description: "Claim-based rules for this identity.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"claim_type": schema.StringAttribute{
										Required: true,
									},
									"claim_value": schema.StringAttribute{
										Required: true,
									},
									"match_type": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ApiPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ApiPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, d := buildApiPolicySave(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.CreateApiPolicy(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Creating API Policy", err.Error())
		return
	}

	plan.ID = types.Int64Value(policy.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetApiPolicy(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading API Policy", err.Error())
		return
	}
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(policy.Name)
	state.Description = types.StringValue(policy.Description)
	state.AllowPublicAccess = types.BoolValue(policy.AllowPublicAccess)
	state.ApiKeyName = types.StringValue(policy.ApiKeyName)
	state.ApiKeyLocation = types.StringValue(policy.ApiKeyLocation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApiPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var state ApiPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	body, d := buildApiPolicySave(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.UpdateApiPolicy(ctx, state.ID.ValueInt64(), body); err != nil {
		resp.Diagnostics.AddError("Updating API Policy", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteApiPolicy(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting API Policy", err.Error())
	}
}

func (r *ApiPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}
	policy, err := r.client.GetApiPolicy(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing API Policy", err.Error())
		return
	}
	if policy == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("API policy %d not found", id))
		return
	}

	emptyList, d := types.ListValue(types.ObjectType{AttrTypes: targetEndpointAttrTypes}, []attr.Value{})
	resp.Diagnostics.Append(d...)
	emptyIdentities, d2 := types.ListValue(types.ObjectType{AttrTypes: identityAttrTypes}, []attr.Value{})
	resp.Diagnostics.Append(d2...)
	emptyTags, d3 := types.ListValue(types.StringType, []attr.Value{})
	resp.Diagnostics.Append(d3...)

	state := ApiPolicyModel{
		ID:                types.Int64Value(policy.ID),
		Name:              types.StringValue(policy.Name),
		Description:       types.StringValue(policy.Description),
		AllowPublicAccess: types.BoolValue(policy.AllowPublicAccess),
		ApiKeyName:        types.StringValue(policy.ApiKeyName),
		ApiKeyLocation:    types.StringValue(policy.ApiKeyLocation),
		Tags:              emptyTags,
		TargetEndpoints:   emptyList,
		Identities:        emptyIdentities,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func buildApiPolicySave(ctx context.Context, model ApiPolicyModel) (client.ApiPolicySave, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	var endpoints []apiPolicyTargetEndpointTF
	if !model.TargetEndpoints.IsNull() && !model.TargetEndpoints.IsUnknown() {
		diagnostics.Append(model.TargetEndpoints.ElementsAs(ctx, &endpoints, false)...)
	}

	var identities []apiPolicyIdentityTF
	if !model.Identities.IsNull() && !model.Identities.IsUnknown() {
		diagnostics.Append(model.Identities.ElementsAs(ctx, &identities, false)...)
	}

	var tags []string
	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		diagnostics.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	}

	if diagnostics.HasError() {
		return client.ApiPolicySave{}, diagnostics
	}

	apiEndpoints := make([]client.ApiPolicyTargetEndpointSave, 0, len(endpoints))
	for _, e := range endpoints {
		apiEndpoints = append(apiEndpoints, client.ApiPolicyTargetEndpointSave{
			URL:    e.URL.ValueString(),
			Method: e.Method.ValueString(),
		})
	}

	apiIdentities := make([]client.ApiPolicyIdentitySave, 0, len(identities))
	for _, ident := range identities {
		var rules []apiPolicyIdentityRuleTF
		if !ident.Rules.IsNull() && !ident.Rules.IsUnknown() {
			diagnostics.Append(ident.Rules.ElementsAs(ctx, &rules, false)...)
		}
		apiRules := make([]client.ApiPolicyIdentityRuleSave, 0, len(rules))
		for _, rule := range rules {
			apiRules = append(apiRules, client.ApiPolicyIdentityRuleSave{
				ClaimType:  rule.ClaimType.ValueString(),
				ClaimValue: rule.ClaimValue.ValueString(),
				MatchType:  rule.MatchType.ValueString(),
			})
		}
		apiIdentities = append(apiIdentities, client.ApiPolicyIdentitySave{
			Name:  ident.Name.ValueString(),
			Rules: apiRules,
		})
	}

	return client.ApiPolicySave{
		Name:              model.Name.ValueString(),
		Description:       model.Description.ValueString(),
		AllowPublicAccess: model.AllowPublicAccess.ValueBool(),
		ApiKeyName:        model.ApiKeyName.ValueString(),
		ApiKeyLocation:    model.ApiKeyLocation.ValueString(),
		Tags:              tags,
		TargetEndpoints:   apiEndpoints,
		Identities:        apiIdentities,
	}, diagnostics
}
