# Roles and Permissions

GPM uses roles for administrative functions and per-process access for everyday users.

## Roles

| Role | Typical use |
| --- | --- |
| Root | Full system administration, including settings, users, permissions, and process definitions. |
| Admin | Daily operations: tasks, notifications, events, metrics, and terminal sharing. |
| User | Access only to processes and logs explicitly assigned by an administrator. |
| Guest | Very limited access. |

Use an admin account for normal operations. Reserve root for security-sensitive changes.

## Process access

An administrator can decide, per process, whether a user may:

- see the process;
- start or stop it;
- open its terminal;
- type into its terminal;
- view its historical logs.

Give the smallest set of access rights needed for the user's work. For example, a support user may need log viewing and a read-only terminal, but not process stop permission.

## Administration tips

- Review user access after team or service ownership changes.
- Revoke terminal-share links when they are no longer needed.
- Use separate accounts instead of sharing administrator credentials.
- Test access with a non-administrator account before handing it to a user.
