# How it works

The provisioner tries to follow to the [spec](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-provisioning.md). It does not have a deal with any remote storage APIs. All it does for provisioning is:
* creates properly named directory under the path specified by `--storage-asset-root` command line interface (_CLI_) flag.
* creates _hostpath_ or _nfs_ type PV which points to specified place of host OS filesystem in the first case or to remote filesytem in the second case. During creating the PV automatically binds to the requesting PVC.

## Details
### CLI-flags

```bash
./provisioner serve --help
starts the watching process for provision/deprovision persistentVolume of K8s

Usage:
  provisioner serve [flags]

Flags:
  -h, --help                        help for serve
      --storage-asset-root string   directory where assets will be created  (requred)
      --storage-classes string      comma separated list of storage class names to watch for (requred)

Global Flags:
  -c, --kubectl-config string   path to kubectl's config
      --v int                   logging verbosity (0..2)

```

### Initialization stage

1. When the provisioner is being started it reads its command line arguments in order to gather input information of further working. These are the list of flags and their meaning:
    * `--storage-classes` - specifies classes name that will be served by the provisioner. The value might be single name or comma separated list of names, for example: _class1,class2,class3_. If at least one of specified class name will not be found the further run will be stopped.
    * `--storage-asset-root` - specifies what directory the provisioner should use as root to create so called `storage asset` for PV.
    * `--kubectl-config` - (optional) specifies path to configuration file for kubectl client library. If it is omitted that it's assumed the provisioner runs inside a cluster.
    * `--v` - (optional) specifies logging level. It might have value from range `0..3`. Default value is 0.
2. Based on input argument's data the provisioner tries to connect to the cluster and get parameters of specified storage classes. Each storage class the provisioner working with must have following keys in `parameters` map:
    * `assetRoot` that is similar of the `--storage-asset-root` CLI-flag. These 2 parameters point to the same place on the shared file system. But the first one is used during creating PV object and for mounting particular PV to a pod by the K8S' controller. The second one is used by only provisioner itself to create a storage asset by OS's syscall and therefore the second path must be mounted into provisioner's pod, if it's supposed to work inside the cluster. But if the provisioner should work outside of the cluster the values of `assetRoot` of the storage class and `--storage-asset-root` of CLI-flag might be the same.
    * `defaultOwnerAssetUid` that is used for set up UID ownership for created storage asset if it is not overridden by `storage.asset/owner-uid` (or `storage-asset.pv.provisioner/owner-uid`) PVC annotation.
    * `defaultOwnerAssetGid` that is used for set up GID ownership for created storage asset if it is not overridden by `storage.asset/owner-gid` (or `storage-asset.pv.provisioner/owner-gid`) PVC annotation.
3. After that it gets started to cycle to watch for:
    * PVCs which need provisioned PVs. It is named `PV provisioning stage`
    * PVs that have been already released and may be deleted. It is named `PV deprovisioning stage`.

### PV provisioning stage

1. In order to determine a PVC that needs new created _PV_ the few conditions should be satisfied. The actual checklist can be found in file [pvc_checkers.go](../cmd/provisioner/checker/pvc_checkers.go). The PVC:
    * must NOT be bound to any _PV_
    * must have `volume.beta.kubernetes.io/storage-provisioner` annotation with value equals to the name of actual provisioner. The value of this annotation is set up by a K8S controller which gets it from the `provisioner` parameter of the storage class
    * must have the same storage class as it was specified by `--storage-classes` CLI-flag
    * must NOT have any _Selectors_

    If any of the mentioned conditions does not satisfied, the PVC is skipped and the provisioner is moving on to next one.

2. If the PVC is met to the conditions, the provisioner tries to create the _storage-asset_ (literally directory) under the path specified by `--storage-asset-root` flag. Full path of created storage asset contains from 3 parts of:
    * `--storage-asset-root` of CLI-flags of provisioner
    * storage class name used for the PVC
    * basename of the storage asset

    Naming convention for basename of new storage asset looks like: __namespaceOfPvc__-__nameOfPvc__-__vol__. The _vol_ suffix is constant string but _namespaceOfPvc_ and _nameOfPvc_ are variables values.

    Before creating storage asset the provisioner checks whether the target path exists or not. By default if the storage asset is already presented on filesystem the provisioner stops any other actions with error message for provision of the PV for requesting PVC. However there are cases when it needs to reuse already existing storage assets for instance due to reinstalling K8S cluster. If the PVC has annotation `storage-asset.pv.provisioner/reuse-existing` with `true` or `yes` value then the provisioner will reuse existing storage asset if any. Otherwise it will try to create it.

    The ownership of the new created storage asset is assigned to UID and GID that can be specified by 2 ways:
    * by annotations `storage.asset/owner-uid` (`storage-asset.pv.provisioner/owner-uid`) or/and `storage.asset/owner-gid` (`storage-asset.pv.provisioner/owner-gid`) of the PVC.
    * by `parameters.defaultOwnerAssetUid` or/and `parameters.defaultOwnerAssetGid` of the class storage.
    The PVC's annotations have more precedence. If a particular annotation on the PVC is absent the corresponding parameter of the storage class will be used.

    If the attempts of creating asset or setting up of ownership are failed, the PVC is skipped and the provisioner is moving to next one.

3. If creating or reusing of the storage asset succeeded the provisioner tries to create a PV for corresponding PVC and bind them to each other. The naming convention for PV is the same as for storage asset: __namespaceOfPvc__-__nameOfPvc__-__vol__.

    Depending on whether colon sign is contained or not in `parameters.assetRoot` of the used storage class for PVC, different types of PV will be created. If value of `parameters.assetRoot` has __colon sign__ the path is considered as NFS share address, and therefore _nfs_ type of PV will be used. Otherwise the path is considered as regular folder name and  _hostPath_ type of PV will be used.

    For created PV the `persistentVolumeReclaimPolicy` parameter is set up whether according to `reclaimPolicy` of the used storage class or to `volume.pv.provisioner/reclaim-policy` annotation of the requesting PVC. The annotation has more precedence.

    If the attempt is failed, the PVC is skipped and the provisioner is moving to next one.

    The example of annotations for PVC can be found [here](../test/test_stuff/02_pvc.yml)

### PV deprovisioning stage

1. In order to determine PV that may be deleted the few conditions should be met. The actual checklist can be found in file [pv_checkers.go](../cmd/provisioner/checker/pv_checkers.go). The PV:
    * must have _Realesed state_.
    * must have the annotation "pv.kubernetes.io/provisioned-by" with value specifying the name of actual provisioner gathered from the storage class.
    * must have `PersistentVolumeClaimPolicy` parameter which has value __Delete__

    If any of the mentioned conditions does not satisfied, the PV is skipped and the provisioner is moving to next one.

2. If the PV is met to the conditions, the provisioner tries to delete storage asset that can be deduced by concatenation of values:
    * `--storage-asset-root` of CLI-flags of provisioner.
    * storage class name used for the PV
    * `PersistentVolume.Name` of the PV

    For details look at [pv_handler.go](../cmd/provisioner/controllers/pv/pv_handler.go) file.

    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.

3. If storage asset removal is succeeded the provisioner tries to delete the PV.

    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.
