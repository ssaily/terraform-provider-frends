package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EnvironmentDataSource{}

func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

type EnvironmentDataSource struct {
	client *client.Client
}

type EnvironmentDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	DisplayName  types.String `tfsdk:"display_name"`
	InternalName types.String `tfsdk:"internal_name"`
	AgentGroups  types.List   `tfsdk:"agent_groups"`
}

type agentGroupSummaryTF struct {
	ID          types.Int64  `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	InternalName types.String `tfsdk:"internal_name"`
}

var agentGroupSummaryAttrTypes = map[string]attr.Type{
	"id":            types.Int64Type,
	"display_name":  types.StringType,
	"internal_name": types.StringType,
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Frends Environment by display name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the environment.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name of the environment to look up.",
				Required:    true,
			},
			"internal_name": schema.StringAttribute{
				Description: "Internal name used by Frends for this environment.",
				Computed:    true,
			},
			"agent_groups": schema.ListNestedAttribute{
				Description: "Agent groups belonging to this environment.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"internal_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(req, resp, &d.client)
}

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environments, err := d.client.ListEnvironments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Listing Environments", err.Error())
		return
	}

	targetName := state.DisplayName.ValueString()
	var found *client.EnvironmentGet
	for i := range environments {
		if environments[i].DisplayName == targetName {
			found = &environments[i]
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Environment Not Found",
			fmt.Sprintf("No environment with display_name %q found", targetName))
		return
	}

	state.ID = types.Int64Value(found.ID)
	state.InternalName = types.StringValue(found.InternalName)

	agGroupObjs := make([]attr.Value, 0, len(found.AgentGroups))
	for _, ag := range found.AgentGroups {
		obj, objDiags := types.ObjectValue(agentGroupSummaryAttrTypes, map[string]attr.Value{
			"id":            types.Int64Value(ag.ID),
			"display_name":  types.StringValue(ag.DisplayName),
			"internal_name": types.StringValue(ag.InternalName),
		})
		resp.Diagnostics.Append(objDiags...)
		agGroupObjs = append(agGroupObjs, obj)
	}

	agGroupList, listDiags := types.ListValue(
		types.ObjectType{AttrTypes: agentGroupSummaryAttrTypes},
		agGroupObjs,
	)
	resp.Diagnostics.Append(listDiags...)
	state.AgentGroups = agGroupList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
