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
  - **CreateOrganization**: Create a new org by name.
  - **GetOrganization**: Get org by id.
  - **ListOrganizations**: List orgs with pagination (common.Pagination).
  - **SuspendOrganization**: Set org status to Suspended.

**Organization** message: `id`, `name`, `status` (OrganizationStatus: ACTIVE, SUSPENDED), `created_at`.

## MembershipService

- **Proto**: [backend/proto/membership/membership.proto](../../../backend/proto/membership/membership.proto). Handler: [internal/membership/handler/grpc.go](../../../backend/internal/membership/handler/grpc.go).
- **RPCs**:
  - **AddMember**: Add a user to an org with a role (org_id, user_id, Role).
  - **RemoveMember**: Remove a user from an org.
  - **UpdateRole**: Change a member’s role (org_id, user_id, Role).
  - **ListMembers**: List members of an org with pagination.

**Role** enum: ROLE_OWNER, ROLE_ADMIN, ROLE_MEMBER. **Member** message: `id`, `user_id`, `org_id`, `role`, `created_at`.

Org-admin operations (e.g. AddMember, RemoveMember, UpdateRole, ListMembers for the dashboard) are protected by **RequireOrgAdmin** so only owner or admin of that org can call them. The dashboard Members page uses API routes that call these RPCs; see [Frontend Dashboard](../frontend/dashboard) (Members section).
