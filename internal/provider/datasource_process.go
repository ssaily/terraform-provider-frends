package provider

import (
	"context"
	"fmt"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProcessDataSource{}

func NewProcessDataSource() datasource.DataSource {
	return &ProcessDataSource{}
}

type ProcessDataSource struct {
	client *client.Client
}

type ProcessDataSourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	GUID             types.String `tfsdk:"guid"`
	Version          types.Int64  `tfsdk:"version"`
	Description      types.String `tfsdk:"description"`
	IsDraft          types.Bool   `tfsdk:"is_draft"`
}

func (d *ProcessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process"
}

func (d *ProcessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Frends Process by name. Processes are read-only in Terraform — they are built in Frends Control Panel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Unique identifier of the process version.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the process to look up.",
				Required:    true,
			},
			"guid": schema.StringAttribute{
				Description: "GUID shared across all versions of this process.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Current build version of the process.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the process.",
				Computed:    true,
			},
			"is_draft": schema.BoolAttribute{
				Description: "Whether the process is a draft.",
				Computed:    true,
			},
		},
	}
}

func (d *ProcessDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureDataSource(req, resp, &d.client)
}

func (d *ProcessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProcessDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processes, err := d.client.ListProcesses(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Listing Processes", err.Error())
		return
	}

	targetName := state.Name.ValueString()
	var found *client.ProcessGet
	for i := range processes {
		if processes[i].Name == targetName && !processes[i].IsDeleted && !processes[i].IsDraft {
			found = &processes[i]
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("Process Not Found",
			fmt.Sprintf("No non-draft, non-deleted process named %q found", targetName))
		return
	}

	state.ID = types.Int64Value(found.ID)
	state.GUID = types.StringValue(found.UniqueIdentifier)
	state.Version = types.Int64Value(int64(found.Version))
	state.Description = types.StringValue(found.Description)
	state.IsDraft = types.BoolValue(found.IsDraft)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
