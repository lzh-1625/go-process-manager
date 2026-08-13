# Task Automation

Tasks let GPM automate process operations. You can run a task on a schedule, as a sequence of steps, when a process changes state, or from an external workflow.

## Create a task

When creating a task, choose:

- the process whose state should be checked;
- the process to start or stop;
- the condition that must be true before the action runs;
- whether the action should continue immediately or wait for completion.

A task can start a stopped process, stop a running process, or wait for the selected action to finish. A task already in progress is not started again.

## Schedule a task

Enable scheduling and enter a Cron expression. GPM accepts standard five-part schedules and schedules that include seconds.

Examples:

```text
# Every day at 03:30
30 3 * * *

# Every 10 seconds
*/10 * * * * *
```

Disable a schedule when you want to pause automatic runs. Test a new schedule with a harmless process first.

## Build a task flow

Link tasks into a sequence to model a simple workflow:

```text
Prepare service → Start service → Check follow-up task
```

The next step runs only after the previous step succeeds. Keep each task focused on one action, avoid circular flows, and make steps safe to retry.

## Trigger from a process event

You can start a task when another process starts, stops, enters a warning state, or changes into another monitored state.

Use this for recovery actions, follow-up jobs, or coordinated service startup. Event-triggered tasks should be tested carefully: they can run even when a schedule is paused.

## Trigger from an external workflow

For integrations, enable API triggering for the task and create its trigger key in the task settings. Use the protected trigger URL supplied by GPM from your automation system.

Treat the trigger URL like a password. Do not expose it in public logs, browser history, or shared chat messages. API triggers start the task in the background; check task activity and events to confirm completion.

## Run and troubleshoot

Administrators can list, start, and stop tasks from the web interface or CLI:

```bash
gpm task list
gpm task start 12
gpm task stop 12
```

For troubleshooting, check the task's running state, recent events, the target process state, and logs from the affected process.
