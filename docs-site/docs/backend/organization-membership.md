---
title: Organization and Membership
sidebar_label: Organization and membership
---

# Organization and Membership

This document describes **organizations** (tenants) and **membership** (users in orgs with roles): OrganizationService and MembershipService, protos, handlers, and how they relate to the dashboard and RBAC. For auth and session behavior, see [auth](./auth) and [sessions](./sessions).

**Audience**: Developers working on multi-tenancy, org admin flows, or RBAC.

## Overview

- **Organizations**: Tenants; each has an id, name, and status (Active, Suspended). Defined in [proto/organization/organization.proto](../../../backend/proto/organization/organization.proto).
- **Membership**: A user belongs to an org with a **role**: Owner, Admin, or Member. MembershipService manages add/remove/update-role and list. Org-admin actions (e.g. dashboard Members page) are restricted to **RequireOrgAdmin** (owner or admin); see [platform RBAC](../../../backend/internal/platform/rbac/).

## OrganizationService

- **Proto**: [backend/proto/organization/organization.proto](../../../backend/proto/organization/organization.proto). Handler: [internal/organization/handler/grpc.go](../../../backend/internal/organization/handler/grpc.go).
- **RPCs**:
  - **CreateOrganization**: Create a new org by name and assign the creating user as owner. **Public endpoint** (no authentication required).
  - **GetOrganization**: Get org by id.
  - **ListOrganizations**: List orgs with pagination (common.Pagination).
  - **SuspendOrganization**: Set org status to Suspended.

**Organization** message: `id`, `name`, `status` (OrganizationStatus: ACTIVE, SUSPENDED), `created_at`.

---

### CreateOrganization

Creates a new organization with the given name and automatically assigns the creating user as the owner. The organization is created with `active` status (auto-activated for PoC). This is a **public endpoint** that does not require authentication, allowing newly registered users to create organizations before they can log in.

**Request** (`CreateOrganizationRequest`):
- `name` (string, required): Organization name. Must be non-empty after trimming whitespace.
- `user_id` (string, required): ID of the user creating the organization. The user must exist in the system.

**Response** (`CreateOrganizationResponse`):
- `organization` (Organization): The created organization with generated `id`, `name`, `status` (ACTIVE), and `created_at` timestamp.

**Validation**:
- `name` must be non-empty after trimming whitespace. Returns `InvalidArgument` if empty.
- `user_id` must be non-empty after trimming whitespace. Returns `InvalidArgument` if empty.
- User must exist in the system. Returns `NotFound` if user does not exist.

**Business Logic**:
1. Validates request parameters (`name` and `user_id`).
2. Verifies the user exists by querying the user repository.
3. Generates a unique organization ID (UUID).
4. Creates the organization with:
   - Generated `id`
   - Provided `name`
   - `status` set to `ACTIVE` (auto-activated for PoC)
   - `created_at` set to current UTC timestamp
5. Creates a membership record linking the user to the organization with `role=owner`.
6. Returns the created organization.

**Error Handling**:
- `InvalidArgument` (400): Missing or empty `name` or `user_id`.
- `NotFound` (404): User with the provided `user_id` does not exist.
- `Internal` (500): Database error during organization or membership creation.

**Security Considerations**:
- This endpoint is **public** (no Bearer token required) because users need to create organizations before they can log in and obtain tokens.
- User validation ensures only registered users can create organizations.
- The creating user is automatically assigned the `owner` role, giving them full administrative control.
- **Note**: In the current PoC implementation, organization creation and membership creation are not wrapped in a database transaction. If membership creation fails after organization creation, the organization will remain in the database without an owner. In production, this should be implemented as a transaction to ensure atomicity.

**Example Flow** (two ways to obtain `user_id`):

1. **After registration**: User registers via `AuthService.Register` and receives `user_id`. User (or frontend) calls `CreateOrganization` with `name` and `user_id`. System creates organization and owner membership. User logs in using the returned organization `id` as `org_id`.

2. **From login page (existing or new user)**: User goes to the login page and uses the "Create new" flow. Frontend calls `AuthService.VerifyCredentials` (email, password) to get `user_id`, then calls `CreateOrganization` with `name` and `user_id`. System creates organization and owner membership. Frontend then logs the user in with the new org.

## MembershipService

- **Proto**: [backend/proto/membership/membership.proto](../../../backend/proto/membership/membership.proto). Handler: [internal/membership/handler/grpc.go](../../../backend/internal/membership/handler/grpc.go).
- **RPCs**:
  - **AddMember**: Add a user to an org with a role (org_id, user_id, Role).
  - **RemoveMember**: Remove a user from an org.
  - **UpdateRole**: Change a member’s role (org_id, user_id, Role).
  - **ListMembers**: List members of an org with pagination.

**Role** enum: ROLE_OWNER, ROLE_ADMIN, ROLE_MEMBER. **Member** message: `id`, `user_id`, `org_id`, `role`, `created_at`.

Org-admin operations (e.g. AddMember, RemoveMember, UpdateRole, ListMembers for the dashboard) are protected by **RequireOrgAdmin** so only owner or admin of that org can call them. The dashboard Members page uses API routes that call these RPCs; see [Frontend Dashboard](../frontend/dashboard) (Members section).

---

## Organization Creation Flow

A user may obtain `user_id` from **Register** (after signup) or from **VerifyCredentials** (e.g. when using the login page "Create new" tab). With that `user_id`, they have no organization membership until they create or join an org. To log in, the user must either:

1. **Create a new organization**: Call `OrganizationService.CreateOrganization` with their `user_id` and an organization name. The `user_id` comes from Register or from VerifyCredentials. The system will:
   - Create the organization with `active` status (auto-activated for PoC)
   - Create a membership record assigning the user as `owner`
   - Return the organization `id` which can be used as `org_id` for login

2. **Join an existing organization**: An organization owner or admin can add the user via `MembershipService.AddMember`, then the user can log in with that organization's `id`.

**Auto-Activation Policy**: For the PoC, organizations are automatically activated (`status=ACTIVE`) upon creation. In production, this would typically require platform administrator approval before activation.

**Owner Assignment**: When a user creates an organization, they are automatically assigned the `owner` role, giving them full administrative control over the organization, including the ability to add/remove members, update roles, configure policies, and manage sessions.

**Transaction Considerations**: Currently, organization creation and membership creation are performed sequentially without a database transaction. If membership creation fails after organization creation succeeds, the organization will exist without an owner. This is acceptable for PoC but should be addressed in production with proper transaction handling.
