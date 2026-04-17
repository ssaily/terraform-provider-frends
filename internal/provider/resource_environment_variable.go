package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EnvironmentVariableResource{}
var _ resource.ResourceWithImportState = &EnvironmentVariableResource{}

func NewEnvironmentVariableResource() resource.Resource {
	return &EnvironmentVariableResource{}
}

type EnvironmentVariableResource struct {
	client *client.Client
}

type EnvironmentVariableModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Values      types.List   `tfsdk:"values"`
}

type envVarValueTF struct {
	EnvironmentID types.Int64  `tfsdk:"environment_id"`
	Value         types.String `tfsdk:"value"`
}

var envVarValueAttrTypes = map[string]attr.Type{
	"environment_id": types.Int64Type,
	"value":          types.StringType,
}

// validEnvVarTypes are the allowed type values for environment variables per the Frends API spec.
const (
	envVarTypeString  = "String"
	envVarTypeObject  = "Object"
	envVarTypeArray   = "Array"
	envVarTypeSecret  = "Secret"
	envVarTypeNumber  = "Number"
	envVarTypeBoolean = "Boolean"
)

var validEnvVarTypes = []string{
	envVarTypeString, envVarTypeObject, envVarTypeArray,
	envVarTypeSecret, envVarTypeNumber, envVarTypeBoolean,
}

func (r *EnvironmentVariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

func (r *EnvironmentVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends environment variable schema and its per-environment values.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the environment variable schema.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the environment variable.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of the environment variable: String, Object, Array, Secret, Number, Boolean.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validEnvVarTypes...),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the environment variable.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"values": schema.ListNestedBlock{
				Description: "Per-environment values for this variable.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"environment_id": schema.Int64Attribute{
							Description: "ID of the environment this value applies to.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "JSON-encoded value for this environment.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *EnvironmentVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *EnvironmentVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schema, err := r.client.CreateEnvironmentVariable(ctx, client.EnvironmentVariableSchemaCreate{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Creating Environment Variable", err.Error())
		return
	}

	plan.ID = types.Int64Value(schema.ID)

	// Set per-environment values.
	values, d := expandEnvVarValues(ctx, plan.Values)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, v := range values {
		if err := r.client.SetEnvironmentVariableValue(ctx, schema.ID, v.EnvironmentID, v.Value); err != nil {
			resp.Diagnostics.AddError("Setting Environment Variable Value", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := r.client.GetEnvironmentVariable(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading Environment Variable", err.Error())
		return
	}
	if ev == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(ev.Name)
	state.Type = types.StringValue(ev.Type)

	// Refresh per-environment values from API so drift is detected on next plan.
	if len(ev.Values) > 0 {
		valueObjs := make([]attr.Value, 0, len(ev.Values))
		for _, v := range ev.Values {
			obj, objDiags := types.ObjectValue(envVarValueAttrTypes, map[string]attr.Value{
				"environment_id": types.Int64Value(v.Environment.ID),
				"value":          types.StringValue(fmt.Sprintf("%v", v.Value)),
			})
			resp.Diagnostics.Append(objDiags...)
			valueObjs = append(valueObjs, obj)
		}
		refreshedValues, listDiags := types.ListValue(
			types.ObjectType{AttrTypes: envVarValueAttrTypes}, valueObjs)
		resp.Diagnostics.Append(listDiags...)
		state.Values = refreshedValues
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EnvironmentVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EnvironmentVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID

	// Update per-environment values.
	values, d := expandEnvVarValues(ctx, plan.Values)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, v := range values {
		if err := r.client.SetEnvironmentVariableValue(ctx, state.ID.ValueInt64(), v.EnvironmentID, v.Value); err != nil {
			resp.Diagnostics.AddError("Setting Environment Variable Value", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEnvironmentVariable(ctx, state.ID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Deleting Environment Variable", err.Error())
	}
}

func (r *EnvironmentVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}
	ev, err := r.client.GetEnvironmentVariable(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing Environment Variable", err.Error())
		return
	}
	if ev == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Environment variable %d not found", id))
		return
	}

	emptyValues, d := types.ListValue(
		types.ObjectType{AttrTypes: envVarValueAttrTypes}, []attr.Value{})
	resp.Diagnostics.Append(d...)

	state := EnvironmentVariableModel{
		ID:     types.Int64Value(ev.ID),
		Name:   types.StringValue(ev.Name),
		Type:   types.StringValue(ev.Type),
		Values: emptyValues,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type expandedEnvVarValue struct {
	EnvironmentID int64
	Value         interface{}
}

func expandEnvVarValues(ctx context.Context, list types.List) ([]expandedEnvVarValue, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	if list.IsNull() || list.IsUnknown() {
		return nil, diagnostics
	}

	var values []envVarValueTF
	diagnostics.Append(list.ElementsAs(ctx, &values, false)...)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	result := make([]expandedEnvVarValue, 0, len(values))
	for _, v := range values {
		result = append(result, expandedEnvVarValue{
			EnvironmentID: v.EnvironmentID.ValueInt64(),
			Value:         v.Value.ValueString(),
		})
	}
	return result, diagnostics
}
