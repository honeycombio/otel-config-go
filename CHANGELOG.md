# otel-config-go changelog

## v1.14.0 (2024-02-06)

### Enhancements

- feat: Populate headers from specified environment variables (#99) | @smoyer64
- feat: Add additional tests for handling OTel env vars for signal generic and specific exporter headers (#106) | @MikeGoldsmith

### Maintenance

- maint: update codeowners to pipeline-team (#97) | @JamieDanielson
- maint: update codeowners to pipeline (#96) | @JamieDanielson
- maint(deps): bump go.opentelemetry.io/proto/otlp from 1.0.0 to 1.1.0 (#105) | @Dependabot
- maint(deps): bump go.opentelemetry.io/otel from 1.21.0 to 1.22.0 (#102) | @Dependabot
- maint(deps): bump google.golang.org/grpc from 1.60.1 to 1.61.0 (#103) | @Dependabot
- maint(deps): bump google.golang.org/grpc from 1.59.0 to 1.60.1 (#100) | @Dependabot

## v1.13.1 (2023-12-04)

### Fixes

- fix: remove duplicate secureOption assignment (#93) | @smoyer64
- fix(pipelines): use protocol constants (#92) | @tranngoclam

### Maintenance

- maint: combine otel core and contrib in dependabot groups (#91) | @JamieDanielson

## v1.13.0 (2023-11-22)

### 💥 Breaking Changes 💥

The OpenTelemetry SDK moved the metrics packages into the main SDK packages with the latest release as it's now GA.
If you used the metrics packages, you may need to update your import paths to reflect the new package.

Additionally, the OpenTelenetry SDK's minimum Go version is now 1.20.

### Maintenance

- Bump OTel dependencies (#84) | @MikeGoldsmith
- Add dependency groups for otel and otel contrib packages (#79) | @MikeGoldsmith
- Bump OTel core and contrib packages to latest (#90) | @MikeGoldsmith

## v1.12.1 (2023-09-21)

### Maintenance

- maint: upgrade otel packages to latest (#68) | @JamieDanielson
- maint: add release.yml for auto-generated release notes (#62) | @JamieDanielson

## v1.12.0 (2023-08-15)

### 💥 Breaking Changes 💥

In previous versions, incompatible resource configurations would fail silently.
Now an error is returned so it is clear when configuration is incompatible.

### Enhancements

- feat: return errors from resource.New (#59) | @dstrelau

### Maintenance

- maint: Match semantic convention version to SDK semantic conventions (#60) | @JamieDanielson

## v1.11.0 (2023-07-28)

### Enhancements

- feat: Add WithResourceOption for additional resource configuration (#48) | @martin308, @robbkidd, @vreynolds

### Maintenance

- docs: Update WithExporterProtocol in README.md (#54) | @NicholasGWK
- maint: update ubuntu and collector versions in CI (#53) | @JamieDanielson
- maint(deps): bump google.golang.org/grpc from 1.56.2 to 1.57.0 (#56)
- maint(deps): bump go.opentelemetry.io/proto/otlp from 0.19.0 to 1.0.0 (#55)

## v1.10.0 (2023-05-31)

### 💥 Breaking Changes 💥

Packages for the Metrics API have been moved as the API implementation has stablized in OTel Go v1.16.0.

- `go.opentelemetry.io/otel/metric/global` -> `go.opentelemetry.io/otel`
- `go.opentelemetry.io/otel/metric/instrument` -> `go.opentelemetry.io/otel/metric`

Imports of these packages in your application will need to be updated.

### Fixes

Fix for the breaking change described above where `go.opentelemetry.io/otel/metric/global` cannot be found for otel-config-go.
The dependency update for otel packages in #40 below—thanks, [Justin Burnham](https://github.com/jburnham)!—includes an update to our import of the metrics package.

### Maintenance

- maint(deps): bump github.com/stretchr/testify from 1.8.2 to 1.8.4 (#44) [dependabot](https://github.com/apps/dependabot)
- maint(deps): bump go.opentelemetry.io/otel from 1.15.1 to 1.16.0 (#40) [@jburnham](https://github.com/jburnham)

## v1.9.0 (2023-05-15)

### 💥 Breaking Changes 💥

- maint: drop go 1.18 (#37) | [@vreynolds](https://github.com/vreynolds)

### Fixes

- fix: Don't fatal error when we can return an error (#36) | [@kentquirk](https://github.com/kentquirk)

### Maintenance

- maint: cleanup go versions (#38) | [@vreynolds](https://github.com/vreynolds)
- maint(deps): bump go.opentelemetry.io/otel from 1.14.0 to 1.15.1 (#35)

## v1.8.0 (2023-04-20)

### Renamed to otel-config-go

[How to migrate to `otel-config-go` from `otel-launcher-go`](/README.md#migrating-from-otel-launcher-go-to-otel-config-go)

### 💥 Breaking Changes 💥

- maint: Rename to otelconfig (#28) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)

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
