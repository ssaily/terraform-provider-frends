package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ApiKeyResource{}
var _ resource.ResourceWithImportState = &ApiKeyResource{}

func NewApiKeyResource() resource.Resource {
	return &ApiKeyResource{}
}

type ApiKeyResource struct {
	client *client.Client
}

type ApiKeyModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	EnvironmentID types.Int64  `tfsdk:"environment_id"`
	Value         types.String `tfsdk:"value"`
	Modified      types.String `tfsdk:"modified"`
	Modifier      types.String `tfsdk:"modifier"`
}

func (r *ApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *ApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends API access key for an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Optional name for the API key.",
				Optional:    true,
			},
			"environment_id": schema.Int64Attribute{
				Description: "ID of the environment this API key belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "The API key value (UUID). Sensitive.",
				Computed:    true,
				Sensitive:   true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp of last modification.",
				Computed:    true,
			},
			"modifier": schema.StringAttribute{
				Description: "User who last modified the key.",
				Computed:    true,
			},
		},
	}
}

func (r *ApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.CreateApiKey(ctx, client.ApiKeyCreate{
		Name:          plan.Name.ValueString(),
		EnvironmentID: plan.EnvironmentID.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating API Key", err.Error())
		return
	}

	plan.ID = types.Int64Value(apiKey.ID)
	plan.Value = types.StringValue(apiKey.Value)
	plan.Modified = types.StringValue(apiKey.Modified)
	plan.Modifier = types.StringValue(apiKey.Modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey, err := r.client.GetApiKey(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading API Key", err.Error())
		return
	}
	if apiKey == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(apiKey.Name)
	state.Value = types.StringValue(apiKey.Value)
	state.Modified = types.StringValue(apiKey.Modified)
	state.Modifier = types.StringValue(apiKey.Modifier)
	state.EnvironmentID = types.Int64Value(apiKey.Environment.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateApiKey(ctx, state.ID.ValueInt64(), client.ApiKeyUpdate{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Updating API Key", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Value = state.Value
	plan.Modified = types.StringValue(updated.Modified)
	plan.Modifier = types.StringValue(updated.Modifier)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteApiKey(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting API Key", err.Error())
	}
}

func (r *ApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}
	apiKey, err := r.client.GetApiKey(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing API Key", err.Error())
		return
	}
	if apiKey == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("API key %d not found", id))
		return
	}
	state := ApiKeyModel{
		ID:            types.Int64Value(apiKey.ID),
		Name:          types.StringValue(apiKey.Name),
		EnvironmentID: types.Int64Value(apiKey.Environment.ID),
		Value:         types.StringValue(apiKey.Value),
		Modified:      types.StringValue(apiKey.Modified),
		Modifier:      types.StringValue(apiKey.Modifier),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
