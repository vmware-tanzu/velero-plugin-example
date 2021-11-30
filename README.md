# Velero Example Plugins

![Build Status][1]

This repository contains example plugins for Velero.

## Kinds of Plugins

Velero currently supports the following kinds of plugins:

- **Object Store** - persists and retrieves backups, backup log files, restore warning/error files, restore logs.
- **Volume Snapshotter** - creates snapshots from volumes (during a backup) and volumes from snapshots (during a restore).
- **Backup Item Action** - performs arbitrary logic on individual items prior to storing them in the backup file.
- **Restore Item Action** - performs arbitrary logic on individual items prior to restoring them in the Kubernetes cluster.
- **Delete Item Action** - performs arbitrary logic on individual items prior to deleting them from the backup file.

Velero can host multiple plugins inside of a single, resumable process. The plugins can be of any supported type. See `main.go`.

For more information, please see the full [plugin documentation](https://velero.io/docs/main/overview-plugins/).

## Building the plugins

To build the plugins, run

```bash
$ make
```

To build the image, run

```bash
$ make container
```

This builds an image tagged as `velero/velero-plugin-example:main`. If you want to specify a different name or version/tag, run:

```bash
$ IMAGE=your-repo/your-name VERSION=your-version-tag make container 
```

## Deploying the plugins

To deploy your plugin image to an Velero server:

1. Make sure your image is pushed to a registry that is accessible to your cluster's nodes.
2. Run `velero plugin add <registry/image:version>`. Example with a dockerhub image: `velero plugin add velero/velero-plugin-example`.

## Using the plugins

When the plugin is deployed, it is only made available to use. To make the plugin effective, you must modify your configuration:

Backup storage:

1. Run `kubectl edit backupstoragelocation <location-name> -n <velero-namespace>` e.g. `kubectl edit backupstoragelocation default -n velero` OR `velero backup-location create <location-name> --provider <provider-name>`
2. Change the value of `spec.provider` to enable an **Object Store** plugin
3. Save and quit. The plugin will be used for the next `backup/restore`

Volume snapshot storage:

1. Run `kubectl edit volumesnapshotlocation <location-name> -n <velero-namespace>` e.g. `kubectl edit volumesnapshotlocation default -n velero` OR `velero snapshot-location create <location-name> --provider <provider-name>`
2. Change the value of `spec.provider` to enable a **Volume Snapshotter** plugin
3. Save and quit. The plugin will be used for the next `backup/restore`

Backup/Restore actions:

1. Add the plugin to Velero as described in the Deploying the plugins section.
2. The plugin will be used for the next `backup/restore`.

## Examples

To run with the example plugins, do the following:

1. Run `velero backup-location create  default --provider file` Optional: `--config bucket:<your-bucket>,prefix:<your-prefix>` to configure a bucket and/or prefix directories.
2. Run `velero snapshot-location create example-default --provider example-volume-snapshotter`
3. Run `kubectl edit deployment/velero -n <velero-namespace>`
4. Change the value of `spec.template.spec.args` to look like the following:

```yaml
      - args:
        - server
        - --default-volume-snapshot-locations
        - example-volume-snapshotter:example-default
```

5. Run `kubectl create -f examples/with-pv.yaml` to apply a sample nginx application that uses the example block store plugin. ***Note***: This example works best on a virtual machine, as it uses the host's `/tmp` directory for data storage.
6. Save and quit. The plugins will be used for the next `backup/restore`

## Creating your own plugin project

1. Create a new directory in your `$GOPATH`, e.g. `$GOPATH/src/github.com/someuser/velero-plugins`
2. Copy everything from this project into your new project

```bash
$ cp -a $GOPATH/src/github.com/vmware-tanzu/velero-plugin-example/* $GOPATH/src/github.com/someuser/velero-plugins/.
```

3. Remove the git history

```bash
$ cd $GOPATH/src/github.com/someuser/velero-plugins
$ rm -rf .git
```

4. Adjust the existing plugin directories and source code as needed.

The `Makefile` is configured to automatically build all directories starting with the prefix `velero-`.
You most likely won't need to edit this file, as long as you follow this convention.

If you need to pull in additional dependencies to your vendor directory, just run

```bash
$ make modules
```

[1]: https://github.com/vmware-tanzu/velero-plugin-example/workflows/Continuous%20Integration/badge.svg

