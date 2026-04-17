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

var _ datasource.DataSource = &AgentGroupDataSource{}

func NewAgentGroupDataSource() datasource.DataSource {
	return &AgentGroupDataSource{}
}

type AgentGroupDataSource struct {
	client *client.Client
}

type AgentGroupDataSourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	DisplayName     types.String `tfsdk:"display_name"`
	InternalName    types.String `tfsdk:"internal_name"`
	IsCrossPlatform types.Bool   `tfsdk:"is_cross_platform"`
	Agents          types.List   `tfsdk:"agents"`
	EnvironmentID   types.Int64  `tfsdk:"environment_id"`
}

func (d *AgentGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_group"
}

func (d *AgentGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Frends Agent Group by display name within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the agent group.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name of the agent group to look up.",
				Required:    true,
			},
			"internal_name": schema.StringAttribute{
				Description: "Internal name used by Frends for this agent group.",
				Computed:    true,
			},
			"is_cross_platform": schema.BoolAttribute{
				Description: "Whether the agent group supports cross-platform agents.",
				Computed:    true,
			},
			"agents": schema.ListAttribute{
				Description: "List of agent names in this group.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"environment_id": schema.Int64Attribute{
				Description: "ID of the environment this agent group belongs to. Required to scope the lookup.",
				Required:    true,
			},
		},
	}
}

func (d *AgentGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(req, resp, &d.client)
}

func (d *AgentGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state AgentGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agentGroups, err := d.client.ListAgentGroupsByEnvironment(ctx, state.EnvironmentID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Listing Agent Groups", err.Error())
		return
	}

	targetName := state.DisplayName.ValueString()
	var found *client.AgentGroupGet
	for i := range agentGroups {
		if agentGroups[i].DisplayName == targetName {
			found = &agentGroups[i]
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Agent Group Not Found",
			fmt.Sprintf("No agent group with display_name %q found in environment %d",
				targetName, state.EnvironmentID.ValueInt64()))
		return
	}

	state.ID = types.Int64Value(found.ID)
	state.InternalName = types.StringValue(found.InternalName)
	state.IsCrossPlatform = types.BoolValue(found.IsCrossPlatform)

	agentVals := make([]attr.Value, 0, len(found.Agents))
	for _, a := range found.Agents {
		agentVals = append(agentVals, types.StringValue(a))
	}
	agentList, listDiags := types.ListValue(types.StringType, agentVals)
	resp.Diagnostics.Append(listDiags...)
	state.Agents = agentList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
