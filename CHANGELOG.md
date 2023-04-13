# otel-config-go changelog

## v1.7.0 (2023-04-11)

### Enhancements

- feat: Allow vendors to set default exporter endpoint (#26) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)

### Maintenance

- ci: add go 1.20 to ci (#25) | [@vreynolds](https://github.com/vreynolds)
- maint: add dependabot to project (#22) | [@JamieDanielson](https://github.com/JamieDanielson)
- maint(deps): bump github.com/sethvargo/go-envconfig from 0.8.2 to 0.9.0 (#24)
- maint(deps): bump google.golang.org/grpc from 1.53.0 to 1.54.0 (#23)

## v1.6.0 (2023-03-29)

No changes have been made to the launcher itself since release v0.3.1.
This new version is being used to help with go dependency resolution and conflicts with other packages.
With this new version it should no longer be required to specify the exact version of the launcher to be downloaded.

## v0.3.1 (2023-03-27)

### Fixes

- fix: `launcher.WithSampler` doesn't get passed all the way through (#17) | [@thomasdesr](https://github.com/thomasdesr)

### Maintenance

- maint: bump semconv to 1.18 (#19) | [@pkanal](https://github.com/pkanal)

## v0.3.0 (2023-03-02)

- Improve behavior for handling custom endpoints (#11) | [@JamieDanielson](https://github.com/JamieDanielson)
- Add example and smoke tests (#7) | [@JamieDanielson](https://github.com/JamieDanielson)
- Bump OTel dependenceis to latest (#14) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)
- Add explicit version to go get command (#15) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)

## v0.2.0 (2023-02-01)

- Update opentelemetry-go dependencies to the January 29th release, resolving a downstream runtime error about missing (deprecated) metric types

## v0.1.0 (2023-01-18)

This has been moved from <https://github.com/honeycombio/opentelemetry-go-contrib/pull/400>
