# Configuration Reference

Use the Settings page to adjust how GPM runs. Save a backup before changing security, networking, log storage, or resource limits.

## Service and security

| Setting | Purpose | When to restart |
| --- | --- | --- |
| Service address and port | Controls where the web interface is available. | Restart after changing. |
| Session lifetime | Controls how long users stay signed in. | New sessions use the new value. |
| Minimum password length | Sets the password policy for new or changed passwords. | No restart normally needed. |
| Terminal timeout | Closes inactive terminal sessions after the selected time. | New sessions use the new value. |

## Logs and storage

| Setting | Purpose |
| --- | --- |
| Log level | Controls how much diagnostic information GPM writes. |
| Log storage | Selects local storage, Elasticsearch, or Bleve. |
| Elasticsearch connection | Supplies the address, index, and credentials for an external log service. |
| Log processing capacity | Controls how much log work GPM can process in parallel. |
| Search result behavior | Controls how GPM counts large Elasticsearch result sets. |

Restart after changing the log storage choice or connection settings. Keep external-service credentials private.

## Process, task, and monitoring behavior

| Setting | Purpose |
| --- | --- |
| Restart limit | Limits automatic restarts of a process. |
| Terminal output cache | Controls how much recent terminal output is retained. |
| Performance history and interval | Controls the amount and frequency of process performance data. |
| Stop wait time | Controls how long GPM waits during a graceful stop. |
| Task timeout | Limits how long a task can wait for completion. |
| WebSocket health interval | Controls terminal connection health checks. |
| Long-poll wait time | Controls how long the interface waits for updates. |
| Event retention | Controls how long audit events are kept. |

## Resource limits and web behavior

| Setting | Purpose |
| --- | --- |
| CPU and memory controls | Enables and tunes resource limits for supported Linux processes. |
| Diagnostic tools | Enables performance troubleshooting tools; expose them carefully. |
| Response compression | Reduces response size for browsers. |
| Static-content caching | Improves loading performance for the web interface. |

## Safe change procedure

1. Record the current values or export a backup.
2. Change one group of settings at a time.
3. Restart GPM when prompted or after a major change.
4. Verify that the web interface, process list, and log search still work.
