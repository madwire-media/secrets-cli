# Getting Started
Previous: [Installing](./0-installing.md)

To start working with the secrets CLI, all you need is to put a `secrets.yaml` file at the root of your project. We recommend putting it at the root of a git repository, but it can really go anywhere. Here's an example of what a `secrets.yaml` file looks like:

```yaml
secrets:
- file: foo.yaml
  vault:
    url: https://vault1.example.com/kv/some/secret
    mapping:
      fromData:
        format: yaml

- file: bar.pem
  vault:
    url: https://vault2.example.com/sandbox/a/different/secret
    mapping:
      fromText:
        path: ['pems', 'bar']
```

This example defines two secret files, `foo.yaml` and `bar.pem`.
* `foo.yaml` comes from the Vault instance at `vault1.example.com`, in the K/V v2 secrets engine aptly named `kv`, from the secret `some/secret`. The entire document from Vault is formatted into YAML.
* `bar.pem` comes from the Vault instance at `vault2.example.com`, in the K/V v2 secrets engine named `sandbox`, from the secret `a/different/secret`. In this case only the string value at `.pems.bar` within the JSON Vault document is used, and it's unformatted.

At the moment this project only supports secrets from Vault, and there are only 2 kinds of mappings. Use `fromData` when your secret is some kind of structured data, like JSON or YAML, and use `fromText` when your secret is a raw text value.

Once you have your `secrets.yaml` file ready, run `secrets sync` to sync between the secrets stores and your local filesystem. The secrets CLI keeps track of changes in a local lockfile (which should _absolutely_ be committed to Git), so when secrets change remotely or locally then the CLI can intelligently decide what to do.

Also since this example connects to two different Vault instances, it will need credentials to access both instances. When you run `secrets sync` in a terminal, it will ask you for those credentials and store them locally. (see the [CI/CD](./4-cicd.md#external-auth) docs for non-tty authentication)

Next: [`secrets.yaml`](./2-secrets-yaml.md)
