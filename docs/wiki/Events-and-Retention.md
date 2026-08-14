# Events and Retention

GPM records important process, task, API, and push activity as system events. Events are useful for auditing who performed an action, checking when a process changed state, and correlating related task operations.

[中文文档](zh/事件记录)

## View and filter events

Administrators can open the **System Events** page to view event time, type, name, user, and additional details. The list can be filtered by time range, event type, name, user, and additional key/value pairs.

Events are stored in the `event` table in [`~/.gpm/data.db`](Log-Storage-and-Search), independently of the selected process-log storage option.

## Recorded events

| Event type | When it is recorded | Additional details |
| --- | --- | --- |
| `ProcessStart` | A managed process enters the running state, including an automatic restart. | Restart count. |
| `ProcessStop` | A managed process finishes exiting. This includes requested stops and unexpected exits. | Process start time and time spent stopping. |
| `ProcessWarning` | A managed process enters the warning state, such as after a start failure or reaching the automatic-restart limit. | Warning reason and process start time when available. |
| `ProcessConnect` | A user or shared terminal connection is attached to a managed process. | The event user identifies the connection. |
| `TaskStart` | A task execution begins, whether triggered manually, by a schedule, through an API key, or by a process-state trigger. | Trace ID. |
| `TaskStop` | A task execution finishes. | Trace ID, success flag, and elapsed time. |
| `ApiRequest` | An authenticated `POST`, `PUT`, or `DELETE` request finishes. Both successful and failed modifying requests are recorded. | Request URI, method, and response status. |
| `PushRequest` | GPM attempts an enabled outbound push request. | Substituted values and either the HTTP status code or transport/build error. |

Read-only `GET` requests are not recorded as `ApiRequest` events. An action may produce more than one event; for example, starting a process through the API can create both an `ApiRequest` event and a `ProcessStart` event.

## Automatic cleanup

The `EventStorageTime` setting controls event retention in days. Its default value is `30`.

- At startup, GPM schedules event cleanup when `EventStorageTime` is `0` or greater.
- Cleanup runs every day at `03:00` in the GPM host's local time zone.
- Each run deletes events created earlier than the current time minus the configured number of days.
- A negative value disables the cleanup schedule when GPM starts.
- A value of `0` deletes events older than the cleanup run itself, effectively clearing existing event history each day.

Restart GPM after changing `EventStorageTime`, especially when enabling or disabling cleanup, so the scheduled job matches the saved setting. Cleanup does not run immediately after startup or after a setting change; the new retention threshold is applied at the next scheduled run. Deleted events cannot be recovered unless `data.db` has been backed up.
