# Releasing

- Update `version.go` with version of release.
- If the new release updates the OpenTelemetry SDK and/or Semantic Convention versions, update the `Latest release built with` section in the [README](./README.md).
- Update `CHANGELOG.md` with the changes since the last release. Consider automating with a command such as these two:
  - `git log $(git describe --tags --abbrev=0)..HEAD --no-merges --oneline > new-in-this-release.log`
  - `git log --pretty='%C(green)%d%Creset- %s | [%an](https://github.com/)'`
- Commit changes, push, and open a release preparation pull request for review.
- Once the pull request is merged, fetch the updated `main` branch.
- Apply a tag for the new version on the merged commit (e.g. `git tag -a v1.2.3 -m "v1.2.3"`)
- Push the tag upstream (this will kick off the release pipeline in CI) e.g. `git push origin v1.2.3`
- Ensure that there is a draft GitHub release created as part of CI publish steps.
- Click "generate release notes" in GitHub for full changelog notes and any new contributors
- Publish the GitHub draft release - if it is a prerelease (e.g. beta) click the prerelease checkbox.
