# Process control duration dialog

## Scope

Update the frontend control-acquisition action for a process. The backend
`GET /process/control` endpoint already accepts `uuid` and `time` query
parameters; this work passes the user-selected duration through that existing
contract.

## Interaction

1. Selecting **Get Control** opens a dialog instead of immediately requesting
   control.
2. The dialog contains a required integer duration field. Its initial value is
   `60`, its unit is seconds, and valid values are from `1` through `3600`.
3. The dialog displays: "获取该进程的控制权，踢掉当前连接终端的用户，设置时间内禁止其他用户访问终端进行写入".
4. Invalid durations cannot be submitted. Cancel closes the dialog without an
   API request.
5. Confirming a valid value calls the existing control API with both the
   process UUID and duration. On success, the dialog closes and keeps the
   existing success snackbar behavior.

## Implementation boundary

The change is limited to the process-card UI, its API helper signature, and
locale strings needed for the dialog. It does not change backend control
semantics or any existing user edits in the worktree.

## Verification

Add focused coverage for duration validation and API parameter forwarding when
the existing frontend test setup supports it; then run the focused test, the
frontend build, and `git diff --check`.
