# CLI Reference

Use the command line to operate a local GPM installation. Run `gpm --help` or `gpm <command> --help` to see available help.

Start the GPM service before using management commands.

## Service commands

| Command | What it does |
| --- | --- |
| `gpm run` | Runs GPM in the foreground. |
| `gpm service install` | Registers GPM as a system service. |
| `gpm service start` | Starts the registered service. |
| `gpm service stop` | Stops the registered service. |
| `gpm service restart` | Restarts the registered service. |
| `gpm service uninstall` | Removes the system-service registration. |
| `gpm version` | Shows the installed version. |
| `gpm config reset` | Restores default settings. This overwrites your current settings. |

## Process commands

| Command | What it does |
| --- | --- |
| `gpm process list` | Shows managed processes that you are allowed to view. |
| `gpm process start NAME` | Starts a process. |
| `gpm process stop NAME` | Stops a process. |
| `gpm process exec NAME` | Opens an interactive terminal for a process when you have permission. |

## Task commands

| Command | What it does |
| --- | --- |
| `gpm task list` | Lists configured tasks. |
| `gpm task start ID` | Starts a task manually. |
| `gpm task stop ID` | Stops a running task. |

## Other management commands

| Command | What it does |
| --- | --- |
| `gpm user list` | Lists user accounts. |
| `gpm user delete ACCOUNT` | Deletes a user account. |
| `gpm push list` | Lists notification destinations. |
| `gpm push delete ID` | Deletes a notification destination. |
| `gpm wsshare list` | Lists active terminal-share links. |
| `gpm wsshare delete ID` | Revokes a terminal-share link. |
| `gpm docker CONTAINER` | On Linux, follows and supervises a Docker container. |

For day-to-day administration, the web interface is usually easier. The CLI is useful for server maintenance and automation.
