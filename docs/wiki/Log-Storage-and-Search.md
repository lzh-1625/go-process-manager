# Log Storage and Search

GPM collects output from managed processes so you can search it from the web interface.

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
