package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProcessTemplateResource{}
var _ resource.ResourceWithImportState = &ProcessTemplateResource{}

func NewProcessTemplateResource() resource.Resource {
	return &ProcessTemplateResource{}
}

type ProcessTemplateResource struct {
	client *client.Client
}

type ProcessTemplateModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Version     types.Int64  `tfsdk:"version"`
}

func (r *ProcessTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process_template"
}

func (r *ProcessTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends Process Template that can be used to create standardized processes.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID of the process template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the process template.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the process template.",
				Optional:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the process template.",
				Computed:    true,
			},
		},
	}
}

func (r *ProcessTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ProcessTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProcessTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := r.client.CreateProcessTemplate(ctx, client.ProcessTemplateCreate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating Process Template", err.Error())
		return
	}

	plan.ID = types.StringValue(template.ID)
	plan.Version = types.Int64Value(int64(template.Version))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProcessTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, err := r.client.GetProcessTemplate(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Reading Process Template", err.Error())
		return
	}
	if template == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(template.Name)
	state.Description = types.StringValue(template.Description)
	state.Version = types.Int64Value(int64(template.Version))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProcessTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateProcessTemplate(ctx, client.ProcessTemplateUpdate{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Updating Process Template", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Version = state.Version
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProcessTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProcessTemplate(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Deleting Process Template", err.Error())
	}
}

func (r *ProcessTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	template, err := r.client.GetProcessTemplate(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Importing Process Template", err.Error())
		return
	}
	if template == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Process template %s not found", req.ID))
		return
	}
	state := ProcessTemplateModel{
		ID:          types.StringValue(template.ID),
		Name:        types.StringValue(template.Name),
		Description: types.StringValue(template.Description),
		Version:     types.Int64Value(int64(template.Version)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
