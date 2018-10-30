# Heptio Ark Example Plugins

**Maintainers**: [Heptio][0]

[![Build Status][1]][2]

This repository contains example plugins for Heptio Ark.

## Kinds of Plugins

Ark currently supports the following kinds of plugins:

- **Object Store** - persists and retrieves backups, backup log files, restore warning/error files, restore logs.
- **Block Store** - creates snapshots from volumes (during a backup) and volumes from snapshots (during a restore).
- **Backup Item Action** - performs arbitrary logic on individual items prior to storing them in the backup file.
- **Restore Item Action** - performs arbitrary logic on individual items prior to restoring them in the Kubernetes cluster.

## Building the plugins

To build the plugins, run

```bash
$ make
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

## Deploying the plugins

To deploy your plugin image to an Ark server:

1. Make sure your image is pushed to a registry that is accessible to your cluster's nodes.
2. Run `ark plugin add <image>`, e.g. `ark plugin add gcr.io/heptio-images/ark-plugin-example`

## Using the plugins

***Note***: As of v0.10.0, the Custom Resource Definitions used to define backup and block storage providers have changed. See [the previous docs][3] for using plugins with versions v0.6-v0.9.x.

When the plugin is deployed, it is only made available to use. To make the plugin effective, you must modify your configuration:

Backup storage:

1. Run `kubectl edit backupstoragelocation <location-name> -n <ark-namespace>` e.g. `kubectl edit backupstoragelocation default -n heptio-ark` OR `ark backup-location create <location-name> --provider <provider-name>`
2. Change the value of `spec.provider` to enable an **Object Store** plugin
3. Save and quit. The plugin will be used for the next `backup/restore`

Volume snapshot storage:

1. Run `kubectl edit volumesnapshotlocation <location-name> -n <ark-namespace>` e.g. `kubectl edit volumesnapshotlocation default -n heptio-ark` OR `ark snapshot-location create <location-name> --provider <provider-name>`
2. Change the value of `spec.provider` to enable a **Block Store** plugin
3. Save and quit. The plugin will be used for the next `backup/restore`

## Examples

To run with the example plugins, do the following:

1. Run `ark backup-location create  default --provider file` Optional: `--config bucket:<your-bucket>,prefix:<your-prefix>` to configure a bucket and/or prefix directories.
2. Run `ark snapshot-location create example-default --provider example-blockstore`
3. Run `kubectl edit deployment/ark -n <ark-namespace>`
4. Change the value of `spec.template.spec.args` to look like the following:

```yaml
      - args:
        - server
        - --default-volume-snapshot-locations
        - example-blockstore:example-default
```

5. Run `kubectl create -f ark-blockstore-example/with-pv.yaml` to apply a sample nginx application that uses the example block store plugin. ***Note***: This example works best on a virtual machine, as it uses the host's `/tmp` directory for data storage.
6. Save and quit. The plugins will be used for the next `backup/restore`

## Creating your own plugin project

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
$ dep ensure
```

## Combining multiple plugins in one file

As of v0.10.0, Ark can host multiple plugins inside of a single, resumable process. The plugins can be
of any supported type. See `ark-examples/main.go`


[0]: https://github.com/heptio
[1]: https://travis-ci.org/heptio/ark-plugin-example.svg?branch=master
[2]: https://travis-ci.org/heptio/ark-plugin-example
[3]: https://github.com/heptio/ark-plugin-example/blob/v0.9.x/README.md#using-the-plugins
