# Releasing

- Update `version.go` with version of release.
- Update `CHANGELOG.md` with the changes since the last release.
- Commit changes, push, and open a release preparation pull request for review.
- Once the pull request is merged, fetch the updated `main` branch.
- Apply a tag for the new version on the merged commit (e.g. `git tag -a v1.2.3 -m "v1.2.3"`)
- Push the tag upstream (this will kick off the release pipeline in CI) e.g. `git push origin v1.2.3`
- Open the draft GitHub release created as part of CI publish steps and "generate release notes"
- Publish the github draft release and this will kick off publishing to GitHub.
