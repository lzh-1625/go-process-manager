# Running the Service

GPM can run in a terminal or in the background as a system service.

## Prepare GPM

Download a release binary, make it executable, and place it in a stable directory. If you build it from source, use the build instructions supplied with the project.

On first start, GPM creates its working data under the launching user's home directory. Keep this directory available and backed up.

## Run in the foreground

Use foreground mode for local testing, debugging, or when another platform already supervises the process:

```bash
./gpm run
```

The service stays attached to the terminal. Press `Ctrl+C` to stop it.

## Run as a system service

For servers, register GPM and then start it:

```bash
./gpm service install
./gpm service start
```

Common maintenance commands:

```bash
./gpm service stop
./gpm service restart
./gpm service uninstall
```

Use the same operating-system account for registration, operation, and access to managed process directories. The account needs permission to run GPM and the services you add to it.

## After startup

1. Open the GPM web interface in a browser.
2. Sign in with an administrator account.
3. Create a small test process before adding production services.
4. After changing settings or upgrading GPM, restart the service.

## Operational tips

- Keep the GPM data directory on persistent storage.
- Restrict access to the GPM host and administrator accounts.
- Stop GPM through the service command rather than killing it abruptly whenever possible.
