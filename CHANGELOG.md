# Changelog
## `v1.0.0`
* Initial public commit

## `v1.0.1`
* Automatically add `secrets.lock` file to `.gitignore`

## `v1.0.2`
* Create parent directories for secret files when they don't exist

## `v1.0.3`
* Log a fatal error when there are multiple secrets at the same file path
* Fix several cases where fatal errors returned a `0` exit code
