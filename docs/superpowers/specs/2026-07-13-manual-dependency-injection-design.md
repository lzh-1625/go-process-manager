# Manual Dependency Injection Design

## Goal

Remove Uber Fx from the entire project and replace its dependency graph and
lifecycle handling with explicit Go code. Preserve the existing runtime
behavior and do not add test files.

## Application Composition

`cmd/go_process_manager/main.go` will be the composition root for the server.
It will construct dependencies in this order:

1. database and generated query handle;
2. repositories;
3. log search backend selected by the existing configuration;
4. event bus, log handler, and application logic;
5. middleware and API handlers;
6. Echo router.

All construction will use the existing `New...` functions. The composition
root will retain only the resulting components needed for startup and
shutdown. `internal/app/module.go`, whose only purpose is to declare Fx
providers, will be deleted.

## Lifecycle

The run command will create a context canceled by `SIGINT` or `SIGTERM` and
will wait on that context instead of relying on `fx.App.Run`.

Startup preserves the current sequence: print the banner, start Echo, start
the event-cleaning cron when enabled, initialize managed processes, initialize
task jobs, and consume event-bus events for triggered tasks. An unexpected
Echo startup failure will be logged and will also end the run loop.

Shutdown will use the existing timeout of five seconds plus the configured
process kill wait time. It will stop accepting HTTP requests, stop cron,
terminate managed processes, close the log handler, close the event bus, and
print the shutdown banner. Shutdown errors will be logged without skipping
the remaining cleanup steps.

## Generator Command

`cmd/gen/gen.go` needs only the database. It will construct the database
directly with `repository.NewDB()` and run GORM generation without creating
the rest of the application graph.

## Dependency Cleanup

The direct `go.uber.org/fx` requirement and its now-unused transitive modules
will be removed by tidying the Go module after all Fx imports are gone.

## Verification

No test files will be added. Verification will format changed Go files, search
for remaining Fx references, run the existing Go test suite, and compile the
relevant commands with the default and `slim` build configurations.
