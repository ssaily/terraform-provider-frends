package provider

import (
	"context"
	"os"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure FrendsProvider satisfies the provider.Provider interface.
var _ provider.Provider = &FrendsProvider{}
var _ provider.ProviderWithFunctions = &FrendsProvider{}

// FrendsProvider is the Terraform provider implementation for Frends iPaaS.
type FrendsProvider struct {
	version string
}

// FrendsProviderModel is the schema model for provider configuration.
type FrendsProviderModel struct {
	HostURL types.String `tfsdk:"host_url"`
	Token   types.String `tfsdk:"token"`
}

// New returns a provider factory function for the given version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FrendsProvider{version: version}
	}
}

// Metadata returns provider metadata.
func (p *FrendsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "frends"
	resp.Version = p.version
}

// Schema defines the provider configuration schema.
func (p *FrendsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Frends iPaaS Platform API to manage integration resources.",
		Attributes: map[string]schema.Attribute{
			"host_url": schema.StringAttribute{
				Description: "Base URL of the Frends Platform API (e.g. https://my-company.frends.com). " +
					"Can be set via the FRENDS_HOST_URL environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				Description: "Bearer token for authenticating with the Frends Platform API. " +
					"Can be set via the FRENDS_TOKEN environment variable.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure validates provider configuration and creates the API client.
func (p *FrendsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config FrendsProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostURL := os.Getenv("FRENDS_HOST_URL")
	if !config.HostURL.IsNull() && !config.HostURL.IsUnknown() {
		hostURL = config.HostURL.ValueString()
	}
	if hostURL == "" {
		resp.Diagnostics.AddError(
			"Missing Frends Host URL",
			"Set the host_url provider attribute or the FRENDS_HOST_URL environment variable.",
		)
	}

	token := os.Getenv("FRENDS_TOKEN")
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Frends Token",
			"Set the token provider attribute or the FRENDS_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(hostURL, token)
	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns all managed resource types.
func (p *FrendsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProcessResource,
		NewProcessDeploymentResource,
		NewEnvironmentVariableResource,
		NewApiKeyResource,
		NewApiPolicyResource,
		NewApiSpecificationResource,
		NewPrivateApplicationResource,
		NewProcessTemplateResource,
	}
}

// DataSources returns all data source types.
func (p *FrendsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewEnvironmentDataSource,
		NewAgentGroupDataSource,
		NewProcessDataSource,
	}
}

// Functions returns custom Terraform functions (none defined yet).
func (p *FrendsProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}
