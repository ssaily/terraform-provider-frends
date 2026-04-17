package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/frends/terraform-provider-frends/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProcessResource{}
var _ resource.ResourceWithImportState = &ProcessResource{}

func NewProcessResource() resource.Resource {
	return &ProcessResource{}
}

type ProcessResource struct {
	client *client.Client
}

type ProcessModel struct {
	PackagePath    types.String `tfsdk:"package_path"`
	ImportConflict types.String `tfsdk:"import_conflict"`
	PackageHash    types.String `tfsdk:"package_hash"`
	// Computed
	ID      types.Int64  `tfsdk:"id"`
	GUID    types.String `tfsdk:"guid"`
	Version types.Int64  `tfsdk:"version"`
	Name    types.String `tfsdk:"name"`
}

func (r *ProcessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process"
}

func (r *ProcessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Frends Process by importing a .frends package file. " +
			"Each update re-imports the package, creating a new process version. " +
			"Deleting this resource also undeploys the process from all agent groups.",
		Attributes: map[string]schema.Attribute{
			"package_path": schema.StringAttribute{
				Description: "Path to the .frends package file to import.",
				Required:    true,
			},
			"import_conflict": schema.StringAttribute{
				Description: "Behaviour when the process GUID already exists on Create: " +
					"Error (fail), UseExisting (skip import), NewVersion (update in-place). " +
					"Defaults to NewVersion. Updates always use NewVersion regardless of this setting.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("NewVersion"),
				Validators: []validator.String{
					stringvalidator.OneOf("Error", "UseExisting", "NewVersion"),
				},
			},
			"package_hash": schema.StringAttribute{
				Description: "SHA-256 hash of the package file. Use filesha256(var.package_path) " +
					"so Terraform detects when the file content changes and triggers a re-import.",
				Optional: true,
				Computed: true,
			},
			"id": schema.Int64Attribute{
				Description: "Numeric ID of the current process version. Changes on each update.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"guid": schema.StringAttribute{
				Description: "GUID of the process, stable across all versions. " +
					"Use this as process_guid in frends_process_deployment.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version number of the process.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Process name as read from the imported package.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProcessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	configureResource(req, resp, &r.client)
}

func (r *ProcessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProcessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hash, err := fileHash(plan.PackagePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Hashing Package File", err.Error())
		return
	}

	result, err := r.client.ImportProcess(ctx, plan.PackagePath.ValueString(), plan.ImportConflict.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Importing Process", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.ID)
	plan.GUID = types.StringValue(result.ElementIdentifier)
	plan.Version = types.Int64Value(int64(result.Version))
	plan.Name = types.StringValue(result.Name)
	plan.PackageHash = types.StringValue(hash)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProcessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	process, err := r.client.GetProcess(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Reading Process", err.Error())
		return
	}
	if process == nil || process.IsDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(process.Name)
	state.GUID = types.StringValue(process.UniqueIdentifier)
	state.Version = types.Int64Value(int64(process.Version))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProcessModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state ProcessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hash, err := fileHash(plan.PackagePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Hashing Package File", err.Error())
		return
	}

	result, err := r.client.ImportProcess(ctx, plan.PackagePath.ValueString(), "NewVersion")
	if err != nil {
		resp.Diagnostics.AddError("Re-importing Process", err.Error())
		return
	}

	plan.ID = types.Int64Value(result.ID)
	plan.GUID = types.StringValue(result.ElementIdentifier)
	plan.Version = types.Int64Value(int64(result.Version))
	plan.Name = types.StringValue(result.Name)
	plan.PackageHash = types.StringValue(hash)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProcessModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProcess(ctx, state.GUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Deleting Process", err.Error())
	}
}

func (r *ProcessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, ok := parseImportID(req.ID, resp)
	if !ok {
		return
	}

	process, err := r.client.GetProcess(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Importing Process", err.Error())
		return
	}
	if process == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Process %d not found", id))
		return
	}

	state := ProcessModel{
		ID:             types.Int64Value(process.ID),
		GUID:           types.StringValue(process.UniqueIdentifier),
		Version:        types.Int64Value(int64(process.Version)),
		Name:           types.StringValue(process.Name),
		ImportConflict: types.StringValue("NewVersion"),
		// package_path and package_hash are unknown until the user adds them to config.
		PackagePath: types.StringValue(""),
		PackageHash: types.StringValue(""),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// fileHash returns the hex-encoded SHA-256 of the file at path.
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing %q: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
