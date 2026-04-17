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

var _ resource.Resource = &ApiSpecificationResource{}
var _ resource.ResourceWithImportState = &ApiSpecificationResource{}

func NewApiSpecificationResource() resource.Resource {
	return &ApiSpecificationResource{}
}

type ApiSpecificationResource struct {
	client *client.Client
}

type ApiSpecificationModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	ActiveVersion types.Int64  `tfsdk:"active_version"`
}

func (r *ApiSpecificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_specification"
}

func (r *ApiSpecificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends API Specification definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the API specification.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the API specification.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the API specification.",
				Optional:    true,
			},
			"active_version": schema.Int64Attribute{
				Description: "Currently active version number.",
				Computed:    true,
			},
		},
	}
}

func (r *ApiSpecificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ApiSpecificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApiSpecificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec, err := r.client.CreateApiSpecification(ctx, client.ApiSpecificationCreate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating API Specification", err.Error())
		return
	}

	plan.ID = types.Int64Value(spec.ID)
	plan.ActiveVersion = types.Int64Value(int64(spec.ActiveVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiSpecificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApiSpecificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec, err := r.client.GetApiSpecification(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading API Specification", err.Error())
		return
	}
	if spec == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(spec.Name)
	state.Description = types.StringValue(spec.Description)
	state.ActiveVersion = types.Int64Value(int64(spec.ActiveVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApiSpecificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ApiSpecificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec, err := r.client.UpdateApiSpecification(ctx, state.ID.ValueInt64(), client.ApiSpecificationUpdate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Updating API Specification", err.Error())
		return
	}

	plan.ID = state.ID
	plan.ActiveVersion = types.Int64Value(int64(spec.ActiveVersion))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApiSpecificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApiSpecificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteApiSpecification(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting API Specification", err.Error())
	}
}

func (r *ApiSpecificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}
	spec, err := r.client.GetApiSpecification(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing API Specification", err.Error())
		return
	}
	if spec == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("API specification %d not found", id))
		return
	}
	state := ApiSpecificationModel{
		ID:            types.Int64Value(spec.ID),
		Name:          types.StringValue(spec.Name),
		Description:   types.StringValue(spec.Description),
		ActiveVersion: types.Int64Value(int64(spec.ActiveVersion)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
