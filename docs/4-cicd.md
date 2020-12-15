# CI/CD
Previous: [Secret Classes](./3-secret-classes.md)

The secrets CLI comes with a CI/CD mode that optimizes execution for CI/CD environments. Enable it with the `--cicd` flag on any command. (Note that some commands may not support CI/CD mode, like updating for example)

When you run `secrets sync --cicd`, the CLI will choose to overwrite local files with their remote secret data whenever there's a discrepancy. It will also ignore the `.localsecretclasses` file, requiring you explicitly set classes on the command line every time.

CI/CD mode will also disable all user prompts as well as any local settings.

## External auth files
In CI/CD mode you will need to provide authentication credentials in your own files. The flag `--auth-config` can be specified one or more times to reference JSON files with auth information. You can also do this outside of CI/CD mode, but it's usually not necessary.

### Cheat Sheet

```json
{
    "vault": {
        "<instance domain>": {
            "userpass": {
                "username": "<username>",
                "password": "<password>"
            },
            "appRole": {
                "roleID": "<role ID>",
                "secretID": "<secret ID>"
            },
            "token": "<token>"
        }
    }
}
```

### Format
* `.vault` - *object*, Vault credentials
    * `.*` - *object*, Vault credentials for a particular domain
        * `.userpass` - *optional object*, Userpass auth method for Vault
            * `.username` - *string*
            * `.password` - *string*
        * `.appRole` - *optional object*, AppRole auth method for Vault
            * `.roleID` - *string*
            * `.secretID` - *string*
        * `.token` - *optional string*, token for direct auth with Vault
