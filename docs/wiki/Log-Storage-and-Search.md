# Log Storage and Search

GPM collects output from managed processes so you can search it from the web interface.

## Local storage paths

By default, GPM stores its local data under `~/.gpm`. Here, `~` is the home directory of the operating-system user that runs GPM. If GPM runs as a service, check the service account's home directory rather than the home directory of the user who opens the web interface.

| Path | Type | Purpose |
| --- | --- | --- |
| `~/.gpm/config.json` | File | Stores GPM settings. GPM creates it with default values when no configuration file exists. |
| `~/.gpm/data.db` | File | Stores GPM data such as processes, users, permissions, tasks, and events. When SQLite log storage is selected, process logs are stored here as well. |
| `~/.gpm/log.bleve` | Directory | Stores the local full-text index created when Bleve log storage is selected. |
| `~/.gpm/log.diskqueue` | Directory | Buffers process logs on disk before they are written to the selected log storage. |

Stop GPM before copying these entries for a backup so that the database, index, and pending log queue remain consistent. The `log.bleve` directory is only present after Bleve storage has been used.

## Choose a storage option

| Option | Best for | Considerations |
| --- | --- | --- |
| SQLite | Small installations and simple local operation. | No external service is needed. |
| Elasticsearch | Larger installations or central log search. | Requires an Elasticsearch service and ongoing capacity management. |
| Bleve | Local full-text search without a separate server. | Uses local disk space and should be included in backups. |

Choose the option that matches your expected log volume and operating model. Changing storage does not automatically move old logs, so plan the change and retain any history you need.

## Search logs

Enter one or more words in the log search box:

- Separate words with spaces to require all of them.
- Put a phrase in double quotes to keep it together: `"out of memory"`.
- Prefix a term with `~` to look for it as part of a longer message.
- Prefix a term with `!` to exclude matching messages.

Examples:

```text
error timeout
"out of memory"
~GET /api
error !debug
```

Search results can vary slightly between storage options, especially for full-text and partial-word searches. Test important alerts with representative production-like logs.
