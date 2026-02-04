/**
 * Membership role enum for use in UI and API.
 * Safe to import from client components (no Node/gRPC dependencies).
 */

export const MembershipRole = {
  UNSPECIFIED: 0,
  OWNER: 1,
  ADMIN: 2,
  MEMBER: 3,
} as const;

export type MembershipRoleType = (typeof MembershipRole)[keyof typeof MembershipRole];

/** Proto enum names when gRPC is loaded with enums: String */
const ROLE_STRING_TO_NUM: Record<string, MembershipRoleType> = {
  ROLE_UNSPECIFIED: 0,
  ROLE_OWNER: 1,
  ROLE_ADMIN: 2,
  ROLE_MEMBER: 3,
};

const ROLE_NUMBERS: number[] = Object.values(MembershipRole) as number[];

/**
 * Normalizes role from API (string enum name or number) to numeric MembershipRoleType.
 * Use for binding select value when API may return role as string.
 */
export function normalizeRole(role: unknown): MembershipRoleType {
  if (typeof role === "number" && ROLE_NUMBERS.includes(role)) {
    return role as MembershipRoleType;
  }
  if (typeof role === "string" && role in ROLE_STRING_TO_NUM) {
    return ROLE_STRING_TO_NUM[role];
  }
  return MembershipRole.UNSPECIFIED;
}
