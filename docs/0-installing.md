# Installing
To install, download the latest release for your OS from GitHub, extract it, and move the executable to a folder somewhere in your command-line PATH. On Linux and macOS a good location is `/usr/local/bin`, which you'll need superuser permissions to copy to. On Windows it might be best to create your own bin folder and add it to your PATH system environment variable manually. (If you _really_ want to you can put the executable file somewhere outside of your PATH, you'll just have to call it with its relative or absolute path instead of just `secrets` globally)

And that's it! No package managers, no additional files or libraries, it's an entirely self-contained program.

## Updating
If you have auto-update turned on, your executable will update itself to the latest version within 24 hours of release. Otherwise you can check for updates with the `secrets update` command.

The update process replaces the currently running `secrets` executable with the new version, which may require elevated permissions. In that case, the command will execute itself with sudo on Linux and macOS, or on Windows it will trigger the User Account Control prompt.

If an update is triggered automatically, then the new `secrets` executable will replace the current process with the same arguments and environment variables in order to carry out your original command.

Next: [Getting Started](./1-getting-started.md)
