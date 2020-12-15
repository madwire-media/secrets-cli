# Secret classes
Previous: [`secrets.yaml`](./2-secrets-yaml.md)

Secret classes are a way to categorize secrets and mark them as optional. When a secret has no assigned class, it will always be synced no matter what. Otherwise users have to locally opt-in to syncing particular secret classes. (All classified secrets are ignored by default)

For example, let's say a team of developers need access to a development secret, that secret is unclassified in their `secrets.yaml` file so everyone always has it no matter what. However, let's say there's another secret, only used in production, that only the team lead can even access. That secret is classified as "prod" in their `secrets.yaml`, and only the team lead has opted in to syncing it.

To adjust your local secret class settings, run the `secrets sync` command with the optional classes argument:
```bash
# uses only classes saved in local class file
secrets sync

# adds all classes (overwrites class file settings)
secrets sync +all

# adds the 'foo' and 'bar' classes
secrets sync +foo,+bar

# removes the 'foo' class
secrets sync ,-foo

# adds all classes except 'foo' (overwrites class file settings)
secrets sync +all,-foo

# reset all classes (overwrites class file settings)
secrets sync ,-all
```

Your secret class preferences for a project are saved locally in a `.localsecretclasses` file which is automatically added to your `.gitignore` as well. (These are local, per-project settings that should never be committed.)

Next: [CI/CD](./4-cicd.md)
