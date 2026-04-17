package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PrivateApplicationResource{}
var _ resource.ResourceWithImportState = &PrivateApplicationResource{}

func NewPrivateApplicationResource() resource.Resource {
	return &PrivateApplicationResource{}
}

type PrivateApplicationResource struct {
	client *client.Client
}

type PrivateApplicationModel struct {
	ID                       types.Int64  `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	DefaultTokenLifetimeDays types.Int64  `tfsdk:"default_token_lifetime_days"`
	CustomTokenClaims        types.Map    `tfsdk:"custom_token_claims"`
	Tags                     types.List   `tfsdk:"tags"`
	Modifier                 types.String `tfsdk:"modifier"`
	ModifiedUtc              types.String `tfsdk:"modified_utc"`
}

func (r *PrivateApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_application"
}

func (r *PrivateApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends Private Application used for OAuth token-based access.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the private application.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the application.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the application.",
				Optional:    true,
			},
			"default_token_lifetime_days": schema.Int64Attribute{
				Description: "Default token lifetime in days (1-730).",
				Required:    true,
			},
			"custom_token_claims": schema.MapAttribute{
				Description: "Custom claims to include in generated tokens (key/value map).",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"tags": schema.ListAttribute{
				Description: "Tags for the application.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"modifier": schema.StringAttribute{
				Description: "User who last modified the application.",
				Computed:    true,
			},
			"modified_utc": schema.StringAttribute{
				Description: "Timestamp of last modification.",
				Computed:    true,
			},
		},
	}
}

func (r *PrivateApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *PrivateApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PrivateApplicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	claims, tags, d := expandClaimsAndTags(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CreatePrivateApplication(ctx, client.PrivateApplicationCreate{
		Name:                     plan.Name.ValueString(),
		Description:              plan.Description.ValueString(),
		DefaultTokenLifetimeDays: int32(plan.DefaultTokenLifetimeDays.ValueInt64()),
		CustomTokenClaims:        claims,
		Tags:                     tags,
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating Private Application", err.Error())
		return
	}

	plan.ID = types.Int64Value(app.ID)
	plan.Modifier = types.StringValue(app.Modifier)
	plan.ModifiedUtc = types.StringValue(app.ModifiedUtc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PrivateApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PrivateApplicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetPrivateApplication(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading Private Application", err.Error())
		return
	}
	if app == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(app.Name)
	state.Description = types.StringValue(app.Description)
	state.DefaultTokenLifetimeDays = types.Int64Value(int64(app.DefaultTokenLifetimeDays))
	state.Modifier = types.StringValue(app.Modifier)
	state.ModifiedUtc = types.StringValue(app.ModifiedUtc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PrivateApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PrivateApplicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	claims, tags, d := expandClaimsAndTags(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.UpdatePrivateApplication(ctx, state.ID.ValueInt64(), client.PrivateApplicationUpdate{
		Name:                     plan.Name.ValueString(),
		Description:              plan.Description.ValueString(),
		DefaultTokenLifetimeDays: int32(plan.DefaultTokenLifetimeDays.ValueInt64()),
		CustomTokenClaims:        claims,
		Tags:                     tags,
	})
	if err != nil {
		resp.Diagnostics.AddError("Updating Private Application", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Modifier = types.StringValue(app.Modifier)
	plan.ModifiedUtc = types.StringValue(app.ModifiedUtc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PrivateApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PrivateApplicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePrivateApplication(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting Private Application", err.Error())
	}
}

func (r *PrivateApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}
	app, err := r.client.GetPrivateApplication(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing Private Application", err.Error())
		return
	}
	if app == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Private application %d not found", id))
		return
	}

	emptyTags, d := types.ListValue(types.StringType, []attr.Value{})
	resp.Diagnostics.Append(d...)
	emptyMap, d2 := types.MapValue(types.StringType, map[string]attr.Value{})
	resp.Diagnostics.Append(d2...)

	state := PrivateApplicationModel{
		ID:                       types.Int64Value(app.ID),
		Name:                     types.StringValue(app.Name),
		Description:              types.StringValue(app.Description),
		DefaultTokenLifetimeDays: types.Int64Value(int64(app.DefaultTokenLifetimeDays)),
		CustomTokenClaims:        emptyMap,
		Tags:                     emptyTags,
		Modifier:                 types.StringValue(app.Modifier),
		ModifiedUtc:              types.StringValue(app.ModifiedUtc),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func expandClaimsAndTags(ctx context.Context, model PrivateApplicationModel) (map[string]interface{}, []string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	claims := map[string]interface{}{}
	var tags []string

	if !model.CustomTokenClaims.IsNull() && !model.CustomTokenClaims.IsUnknown() {
		var claimsStr map[string]string
		diagnostics.Append(model.CustomTokenClaims.ElementsAs(ctx, &claimsStr, false)...)
		for k, v := range claimsStr {
			claims[k] = v
		}
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		diagnostics.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	}

	return claims, tags, diagnostics
}
