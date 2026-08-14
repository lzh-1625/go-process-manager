# Platform Differences

GPM runs on multiple operating systems, but some process-management features depend on operating-system facilities. Review these differences before using the same process configuration on another platform.

[中文文档](zh/系统差异)

## Resource limits with cgroup

GPM's cgroup-based CPU and memory limits are available only on Linux.

- The Linux host must provide a usable cgroup v1 or cgroup v2 environment, and the account running GPM must have permission to create and manage cgroups.
- If cgroup support is unavailable or initialization fails, the managed process still starts, but the configured CPU and memory limits do not take effect. Check the GPM log for the reason.
- On Windows and other non-Linux systems, enabling cgroup for a process has no effect.

## Stopping processes

The behavior of the normal **Stop** action differs by platform:

| Platform | Behavior |
| --- | --- |
| Linux | GPM sends `SIGINT` first and waits up to the configured stop wait time. If the process does not exit, GPM forcefully kills it. |
| Windows | Windows does not support sending `SIGINT` through Go's `os.Process.Signal`. GPM therefore falls back immediately to its force-stop (`Kill9`) path without waiting for the configured stop wait time. |

`Kill9` is GPM's internal name for an immediate force stop. On Unix-like systems this corresponds to a forceful process kill such as `SIGKILL`; on Windows, Go terminates the process with the Windows `TerminateProcess` API. Windows does not have a literal `kill -9` signal.

Applications that need graceful shutdown should provide their own platform-appropriate shutdown mechanism, especially when they run under GPM on Windows.
