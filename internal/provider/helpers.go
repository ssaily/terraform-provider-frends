package provider

import (
	"fmt"
	"strconv"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// configureResourceClient extracts the *client.Client from provider data and assigns it.
// Returns false and adds a diagnostic error if provider data is of an unexpected type.
func configureResourceClient(providerData any, target **client.Client, diags interface {
	AddError(string, string)
}) bool {
	if providerData == nil {
		return true
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError("Unexpected Provider Data",
			fmt.Sprintf("Expected *client.Client, got %T", providerData))
		return false
	}
	*target = c
	return true
}

// configureResource is the Configure handler for resource types.
func configureResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse, target **client.Client) {
	configureResourceClient(req.ProviderData, target, &resp.Diagnostics)
}

// configureDataSource is the Configure handler for data source types.
func configureDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse, target **client.Client) {
	configureResourceClient(req.ProviderData, target, &resp.Diagnostics)
}

// parseImportID parses a numeric string ID from an import request.
// Returns the parsed int64 and true on success. On failure, adds a diagnostic error and returns false.
func parseImportID(id string, resp *resource.ImportStateResponse) (int64, bool) {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected a numeric resource ID, got: %q", id))
		return 0, false
	}
	return parsed, true
}
