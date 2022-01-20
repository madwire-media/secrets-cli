# Changelog
## `v1.3.2`
* Fixed binary name in release assets

## `v1.3.1`
* Included macOS .tar.gz archives in GitHub release

## `v1.3.0`
* Added Apple silicon support (darwin_arm64)
* Signed and notarized macOS binaries

## `v1.2.0`
* Update `--pull` and `--push` behavior to be "only X" instead of "X by default"
    * When there are conflicts or errors between local and remote, the flags will still work as normal
    * When there is only a local change, `--pull` will now prevent that change from being pushed
    * Similarly, when there is only a remote change, `--push` will now prevent that change from being pulled
* Fix linting warnings in the code

## `v1.1.1`
* Fix `secrets config login` AppRole auth with credentials supplied as arguments

## `v1.1.0`
* Add OIDC auth support
* Add auth selector UI
* Add `secrets add <file>` command to help with editing the `secrets.yaml` file
* Add `secrets config login [host]` command to interactively configure user and external auth files
* Add `secrets config autoupdate` command to interactively enable or disable automatic updates
* Rename `secrets update` to `secrets self-update`
* Other minor code improvements

## `v1.0.5`
* Fix previous bug but now for CI/CD mode

## `v1.0.4`
* Fix a bug where the token cache may be loaded as null

## `v1.0.3`
* Log a fatal error when there are multiple secrets at the same file path
* Fix several cases where fatal errors returned a `0` exit code

## `v1.0.2`
* Create parent directories for secret files when they don't exist

## `v1.0.1`
* Automatically add `secrets.lock` file to `.gitignore`

## `v1.0.0`
* Initial public commit
