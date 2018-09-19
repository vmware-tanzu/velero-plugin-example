# Example blockstore plugin

This directory contains source code for an ark blockstore plugin, it implements a no-op
plugin ([this API](https://github.com/heptio/ark/blob/283a1349bdec670a1be0f3bfa19b07550c05b88f/pkg/cloudprovider/storage_interfaces.go#L62)).

**It works only for hostpath PersistentVolumes**. Using it with other PersistentVolume types can have
undefined behavior.

## Prerequisites
### To build
- Docker

### To run/test
- A Kubernetes cluster
- Ark (look [here](https://github.com/heptio/ark/blob/master/docs/quickstart.md) for install instructions)


## Building this plugin

1. Check out the sources
  ```bash
  git clone https://github.com/heptio/ark-plugin-example.git
  ```

2. Build the plugin Docker image
  ```bash
  make container
  ```
You should now have a Docker image named  `gcr.io/heptio-images/ark-plugin-example:latest`  
Push it to a Docker registry that your Kubernetes cluster can pull from.

# Deployment
### Installation
```bash
   ark plugin add gcr.io/heptio-images/ark-plugin-example:latest
```

### Configuration
1. ```bash
      kubectl edit config -n heptio-ark
   ```
2. Set up "example" as the persistentVolumeProvider by adding the following snippet to the config spec:
   ```yaml
   persistentVolumeProvider:
      name: example
   ```
   
# Testing the plugin
1. Install Test Application  
   Now install an nginx-application with a pv like this :
   ```bash
   kubectl create -f ark-blockstore-example/pv.yaml
   kubectl create -f ark-blockstore-example/pvc.yaml
   kubectl create -f ark-blockstore-example/with-pv.yaml
   ```

2. Run a backup  
   ```bash
   ark backup create nginx-backup --selector app=nginx
   ```

3. Check backup status  
   ```bash
   ark backup get
   ```

4. Delete backup  
  ```bash
  ark backup delete nginx-backup
  ```


