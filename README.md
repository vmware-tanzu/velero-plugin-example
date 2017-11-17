# Heptio Ark Example Plugins

**Maintainers**: [Heptio][0]

[![Build Status][1]][2]

This repository contains example plugins for Heptio Ark.

# Kinds of Plugins

Ark currently supports the following kinds of plugins:

- **Object Store** - persists and retrieves backups, backup log files, restore warning/error files, restore logs.
- **Block Store** - creates snapshots from volumes (during a backup) and volumes from snapshots (during a restore).
- **Backup Item Action** - performs arbitrary logic on individual items prior to storing them in the backup file.
- **Restore Item Action** - performs arbitrary logic on individual items prior to restoring them in the Kubernetes cluster.

# Building the plugins

To build the plugins, run

```bash
$ make container
```

To build the image, run

```bash
$ make container
```

This builds an image tagged as `gcr.io/heptio-images/ark-plugin-example`. If you want to specify a
different name, run

```bash
$ make container IMAGE=your-repo/your-name:here
```

# Deploying the plugins

To deploy your plugin image to an Ark server:

1. Make sure your image is pushed to a registry that is accessible to your cluster's nodes.
1. Run `ark plugin add <image>`, e.g. `ark plugin add gcr.io/heptio-images/ark-plugin-example`

# Creating your own plugin project

1. Create a new directory in your `$GOPATH`, e.g. `$GOPATH/src/github.com/someuser/ark-plugins`
2. Copy everything from this project into your new project

```bash
$ cp -a $GOPATH/src/github.com/heptio/ark-plugin-example/* $GOPATH/src/github.com/someuser/ark-plugins/.
```

3. Remove the git history

```bash
$ cd $GOPATH/src/github.com/someuser/ark-plugins
$ rm -rf .git
```

4. Adjust the existing plugin directories and source code as needed.

The `Makefile` is configured to automatically build all directories starting with the prefix `ark-`.
You most likely won't need to edit this file, as long as you follow this convention.

If you need to pull in additional dependencies to your vendor directory, just run

```bash
$ ./dep-save.sh
```

# Combining multiple plugins in one file

Note that currently, Ark uses the [name of the plugin binary][3] to determine the type and unique name
of the plugin. This means that Ark will only recognize one plugin per binary file.

If you want to implement more than one plugin in a single binary, you can create hard or symbolic
links to your binary and add them to the image, changing the name of each link to match the name of
the desired plugin.

For example, if your binary is `ark-awesome-plugins`, you could create hard/symoblic links
`ark-objectstore-mycloud` and `ark-blockstore-mycloud` that both point to `ark-awesome-plugins`.

In the future, we do hope to make it easier to register a multi-plugin binary with Ark - stay tuned!

[0]: https://github.com/heptio
[1]: https://travis-ci.org/heptio/ark-plugin-example.svg?branch=master
[2]: https://travis-ci.org/heptio/ark-plugin-example
[3]: https://github.com/heptio/ark/blob/master/docs/plugins.md#plugin-naming
