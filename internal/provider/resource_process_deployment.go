package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProcessDeploymentResource{}
var _ resource.ResourceWithImportState = &ProcessDeploymentResource{}

func NewProcessDeploymentResource() resource.Resource {
	return &ProcessDeploymentResource{}
}

type ProcessDeploymentResource struct {
	client *client.Client
}

type ProcessDeploymentModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	AgentGroupID          types.Int64  `tfsdk:"agent_group_id"`
	Processes             types.List   `tfsdk:"processes"`
	ActivateTriggers      types.Bool   `tfsdk:"activate_triggers"`
	DeploymentDescription types.String `tfsdk:"deployment_description"`
	TriggersActive        types.Bool   `tfsdk:"triggers_active"`
}

type processVersionTF struct {
	ProcessGUID      types.String `tfsdk:"process_guid"`
	Version          types.Int64  `tfsdk:"version"`
	ProcessVariables types.List   `tfsdk:"process_variables"`
}

type processVariableTF struct {
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	IsSecret    types.Bool   `tfsdk:"is_secret"`
	Description types.String `tfsdk:"description"`
}

var processVariableAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"value":       types.StringType,
	"is_secret":   types.BoolType,
	"description": types.StringType,
}

var processVersionAttrTypes = map[string]attr.Type{
	"process_guid": types.StringType,
	"version":      types.Int64Type,
	"process_variables": types.ListType{
		ElemType: types.ObjectType{AttrTypes: processVariableAttrTypes},
	},
}

func (r *ProcessDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process_deployment"
}

func (r *ProcessDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends process deployment. Changing agent_group_id or processes forces replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the process deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"agent_group_id": schema.Int64Attribute{
				Description: "ID of the Agent Group to deploy to. Changing this forces replacement.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"activate_triggers": schema.BoolAttribute{
				Description: "Whether to activate Triggers on deployment. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"deployment_description": schema.StringAttribute{
				Description: "Optional description for this deployment.",
				Optional:    true,
			},
			"triggers_active": schema.BoolAttribute{
				Description: "Current trigger activation state.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"processes": schema.ListNestedBlock{
				Description: "Processes to deploy. Changing this forces replacement.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"process_guid": schema.StringAttribute{
							Description: "GUID of the process.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"version": schema.Int64Attribute{
							Description: "Build version of the process.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"process_variables": schema.ListNestedBlock{
							Description: "Process-level variable overrides.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Optional:  true,
										Sensitive: false,
									},
									"is_secret": schema.BoolAttribute{
										Optional: true,
										Computed: true,
									},
									"description": schema.StringAttribute{
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

func (r *ProcessDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ProcessDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProcessDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processes, d := expandProcessVersions(ctx, plan.Processes)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.CreateProcessDeployment(ctx, client.ProcessDeploymentCreate{
		AgentGroupID:          plan.AgentGroupID.ValueInt64(),
		Processes:             processes,
		ActivateTriggers:      plan.ActivateTriggers.ValueBool(),
		DeploymentDescription: plan.DeploymentDescription.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating Process Deployment", err.Error())
		return
	}

	plan.ID = types.Int64Value(deployment.DeploymentID)
	plan.TriggersActive = types.BoolValue(deployment.TriggersActive)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProcessDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.GetProcessDeployment(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading Process Deployment", err.Error())
		return
	}
	if deployment == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.TriggersActive = types.BoolValue(deployment.TriggersActive)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProcessDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ActivateTriggers.Equal(state.ActivateTriggers) {
		if err := r.client.SetDeploymentActivation(ctx, state.ID.ValueInt64(), plan.ActivateTriggers.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Updating Deployment Activation", err.Error())
			return
		}
	}

	plan.ID = state.ID
	plan.TriggersActive = plan.ActivateTriggers
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProcessDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProcessDeployment(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting Process Deployment", err.Error())
	}
}

func (r *ProcessDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected numeric deployment ID, got: %s", req.ID))
		return
	}
	deployment, err := r.client.GetProcessDeployment(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing Process Deployment", err.Error())
		return
	}
	if deployment == nil {
		resp.Diagnostics.AddError("Not Found",
			fmt.Sprintf("Deployment %d not found", id))
		return
	}

	emptyProcesses, d := types.ListValue(
		types.ObjectType{AttrTypes: processVersionAttrTypes},
		[]attr.Value{},
	)
	resp.Diagnostics.Append(d...)

	state := ProcessDeploymentModel{
		ID:             types.Int64Value(deployment.DeploymentID),
		AgentGroupID:   types.Int64Value(deployment.AgentGroup.ID),
		TriggersActive: types.BoolValue(deployment.TriggersActive),
		Processes:      emptyProcesses,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func expandProcessVersions(ctx context.Context, list types.List) ([]client.ProcessVersionInput, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diagnostics
	}

	var versions []processVersionTF
	diagnostics.Append(list.ElementsAs(ctx, &versions, false)...)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	result := make([]client.ProcessVersionInput, 0, len(versions))
	for _, v := range versions {
		vars, d := expandProcessVariables(ctx, v.ProcessVariables)
		diagnostics.Append(d...)
		if diagnostics.HasError() {
			return nil, diagnostics
		}
		result = append(result, client.ProcessVersionInput{
			ProcessGUID:      v.ProcessGUID.ValueString(),
			Version:          int32(v.Version.ValueInt64()),
			ProcessVariables: vars,
		})
	}
	return result, diagnostics
}

func expandProcessVariables(ctx context.Context, list types.List) ([]client.ProcessVariable, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diagnostics
	}

	var vars []processVariableTF
	diagnostics.Append(list.ElementsAs(ctx, &vars, false)...)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	result := make([]client.ProcessVariable, 0, len(vars))
	for _, v := range vars {
		result = append(result, client.ProcessVariable{
			Name:        v.Name.ValueString(),
			Value:       v.Value.ValueString(),
			IsSecret:    v.IsSecret.ValueBool(),
			Description: v.Description.ValueString(),
		})
	}
	return result, diagnostics
}
