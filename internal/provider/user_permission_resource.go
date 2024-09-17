// Copyright (c) Predell Services
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserPermissionResource() resource.Resource {
	return &UserPermissionResource{}
}

// UserPermissionResource defines the resource implementation.
type UserPermissionResource struct {
	client arangodb.Client
}

// UserPermissionResourceModel describes the resource data model.
type UserPermissionResourceModel struct {
	Database   types.String `tfsdk:"database"`
	Permission types.String `tfsdk:"permission"`
	User       types.String `tfsdk:"user"`
}

func (r *UserPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permission"
}

func (r *UserPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A user permission to access a database",

		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "Permission to access the database, can be 'ro' for read only, 'rw' for read-write or 'none'",
				Required:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The name of the user",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
		},
	}
}

func (r *UserPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*arangodb.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *arangodb.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = *client
}

func (r *UserPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserPermissionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, errGetUser := r.client.User(ctx, data.User.ValueString())
	if errGetUser != nil {
		resp.Diagnostics.AddError(
			"Unable to find existing User",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+errGetUser.Error(),
		)

		return
	}

	err := user.SetDatabaseAccess(ctx, data.Database.ValueString(), arangodb.Grant(data.Permission.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			"An unexpected error occurred while attempting to create the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserPermissionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.User(ctx, data.User.ValueString())

	if err != nil {
		if shared.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to find existing User",
				"An unexpected error occurred while attempting to refresh resource state. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Error: "+err.Error(),
			)
		}
		return
	}

	access, err := user.GetDatabaseAccess(ctx, data.Database.ValueString())
	if err != nil {
		if shared.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				"Unable to get permissions",
				"An unexpected error occurred while attempting to create the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Error: "+err.Error(),
			)
		}
		return
	}

	data.Permission = types.StringValue(string(access))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserPermissionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, errUser := r.client.User(ctx, data.User.ValueString())

	if errUser != nil {
		resp.Diagnostics.AddError(
			"Unable to get User",
			"An unexpected error occurred while attempting to update the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+errUser.Error(),
		)

		return
	}

	errDatabase := user.SetDatabaseAccess(ctx, data.Database.ValueString(), arangodb.Grant(data.Permission.ValueString()))
	if errDatabase != nil {
		if shared.IsNotFound(errDatabase) {
			resp.Diagnostics.AddError(
				"Unable to find database",
				"An unexpected error occurred while attempting to update the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Error: "+errDatabase.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				"An unexpected error occurred while attempting to update the resource. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"HTTP Error: "+errDatabase.Error(),
			)
		}
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserPermissionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, errUser := r.client.User(ctx, data.User.ValueString())

	if errUser != nil {
		resp.Diagnostics.AddError(
			"Unable to get existing user",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+errUser.Error(),
		)

		return
	}

	err := user.RemoveDatabaseAccess(ctx, data.Database.ValueString())
	if err != nil && !shared.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}
}

func (r *UserPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
