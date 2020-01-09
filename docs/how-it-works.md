# How it works

The provisioner tries to follow to the [spec](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-provisioning.md). It does not have a deal with any remote storage APIs. All it does for provisioning is
* creates properly named directory under the path specified by `--storage-asset-root` command line interface (_CLI_) flag, which is mounted from the host OS into the provisioner's container
* creates _hostpath_ PV which points into the specified directory.

### Details
#### Initialization stage

1. When the provisioner starts it reads from CLI-flags in order to know what kind of unbound PVCs it should watch to create PV for them. The meaning flag for that is `--storage-class`.
2. From the CLI-flags the provisioner discovers what directory it should use as root to create so called `storage asset` of PV. The meaning flag is `--storage-asset-root`
3. Another CLI flag says to provisioner whether it runs inside or outside of a K8s cluster. The meaning flag is `--kubectl-config`. This flag is optional and if it is omitted that it's assumed running the provisioner inside a cluster.
4. Based on input argument's data the provisioner tries to connect to the cluster and get parameters of specified storage class. One of the storage class parameter is `assetRoot` that is similar of the `--storage-asset-root` CLI-flag. These 2 parameters point to the same place on the shared file system. But the first one is used during creating PV object and for mounting particular PV to a pod by the K8s controller. The second one is used by only provisioner itself to create a storage asset by OS's syscall and therefore the second path must be mounted into provisioner's pod  (if it's supposed to work inside the cluster). But if the provisioner should work outside of the cluster the values of `assetRoot` of the storage class and `--storage-asset-root` of CLI-flag might be the same.
5. After that it gets started to cycle to watch for:
    * PVCs which need provisioned PVs. It is named `PV provisioning stage`
    * PVs that have been already released and may be deleted. It is named `PV deprovisioning stage`.

#### PV provisioning stage
1. In order to determine a PVC that needs new created _PV_ the few conditions should be satisfied. The PVC:
    * should have "volume.beta.kubernetes.io/storage-provisioner" annotation with value equals to the name of actual provisioner. This value is set up by a controller, which is looking for it in the storage class
    * should have the same storage class as it was specified by `--storage-class` CLI-flag
    * must not be bound to any _PV_
    * must not have any _Selectors_

    If any of the mentioned condition does not satisfied, the PVC is skipped and the provisioner is moving on to next one.
2. If the PVC is met to the conditions, the provisioner tries to create the _storage-asset_ (literally directory) under the path specified by `--storage-asset-root` flag. Naming convention of the new storage asset looks like: __namespaceOfPvc__-__nameOfPvc__-__vol__. The _vol_ suffix is constant string but _namespaceOfPvc_ and _nameOfPvc_ are variables values.

    The ownership of the new created storage asset is assigned to UID and GID that can be specified by 2 ways:
    * by annotations `storage.asset/owner-uid` or/and `storage.asset/owner-gid` for PVC.
    * by parameters `defaultOwnerAssetUid` or/and `defaultOwnerAssetGid` of the class storage.
    The PVC's annotations have more weight. If a particular annotation on the PVC is absent the corresponding parameter of the storage class will be used. To clarify see how the `pvc06` looks like in the _test_stuff_ catalog.

    If the attempt of creating asset or setting of ownership are failed, the PVC is skipped and the provisioner is moving to the one.
3. If creating of the storage asset succeeded the provisioner tries to create a PV for corresponding PVC and bind them to each other. The naming convention for PV is the same as for storage asset: __namespaceOfPvc__-__nameOfPvc__-__vol__.

    If the attempt is failed, the PVC is skipped and the provisioner is moving to next one.

#### PV deprovisioning stage
1. In order to determine PV that may be deleted the few conditions should be met. The PV
    * should have the annotation "pv.kubernetes.io/provisioned-by" with value specifying the name of actual provisioner gathered from the storage class.
    * should have `PersistentVolumeClaimPolicy` parameter which has value __Delete__

    If any of the mentioned conditions does not satisfied, the PV is skipped and the provisioner is moving to next one.
2. If the PV is met to the conditions, the provisioner tries to delete storage asset that can be deduced by parameters compilation like `PersistentVolumeSource.HostPath.Path` of the PV and `--storage-asset-root` of CLI-flags of provisioner.

    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.

3. If storage asset deleting is succeeded the provisioner tries to delete the PV.

    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.
