# `secrets.yaml`
Previous: [Getting started](./1-getting-started.md)

## Cheat Sheet
```yaml
secrets:
  - file: <file path>
    class: <class> # optional
    vault: # optional
      url: <url to Vault secret>
      mapping:
        fromData: # optional
          format: <format>
          path: ['<key 1>', '<key 2>', '...'] # optional
        fromText: # optional
          path: ['<key 1>', '<key 2>', '...']
```

## Structure
### Root
**Object**
* `.secrets` - *array of [Secret]*, list of secrets

### Secret
**Object**
* `.file` - *string*, local path where secret should be stored
* `.class` - *optional string*, classification of secret (see [Secret Classes](./3-secret-classes.md))
* `.vault` - *optional [VaultSecret]*, configuration to sync this secret with Vault

### VaultSecret
**Object**
* `.url` - *string*, URL to the Vault secret in the format of `http[s]://<domain>/<engine>/<secret path>`
* `.mapping` - *object*
    * `.fromData` - *optional [VaultDataMapping]*, maps this secret to structured data in Vault
    * `.fromText` - *optional [VaultTextMapping]*, maps this secret to text data in Vault

### VaultDataMapping
**Object**
* `.format` - *[DataFormat]*, format to render the local secret as
* `.path` - *optional array of string or int*, list of map or array indexes to secret data in a Vault document

### VaultTextMapping
**Object**
* `.path` - *array of string or int*, list of map or array indexes to string in a Vault document

### DataFormat
**Enum**
One of:
* `json` - format as JSON
* `yaml` - format as YAML

Next: [Secret Classes](./3-secret-classes.md)

[Secret]: #secret
[VaultSecret]: #vaultsecret
[VaultDataMapping]: #vaultdatamapping
[VaultTextMapping]: #vaulttextmapping
[DataFormat]: #dataformat
