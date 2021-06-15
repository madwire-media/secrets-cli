# Changelog
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
