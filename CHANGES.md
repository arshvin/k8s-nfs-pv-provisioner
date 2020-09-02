# Change list
* 0.4.1 - Updated docs according to changes. Removed excess iota statements.
* 0.4.0 - Implemented choosing of `nfs` or `hostPath` type volume depending on presence of colon sign in `assetRoot` of a storage class
* 0.3.2 - Cleaning code in order to remove code duplication for starting controllers
* 0.3.1 - Cleaning code in order to reworking PVC/PV checker approach
* 0.3.0 - Added the feature of reusing existing storage asset if there is annotation `storage-asset.pv.provisioner/reuse-existing` on PVC
* 0.2.1 ... 0.2.6 - Bug fixes
* 0.2.0 - Added:
  1. ability to pass more than one storage class name through CLI flag as comma separated list of names.
  2. processing of `ReclaimPolicy` of the storage class. Earlier the "Delete" reclaim policy was assigned to PV as hard coded value.
  2. processing of `pv-provisioner/reclaim-policy` on PVC.
* 0.1.3 - Fully reworked algorithm of the provisioner to use Working queues and Informers.
* 0.1.2 - Added support of `storage.asset/owner-uid` and `storage.asset/owner-gid` annotations for PVC.
* 0.1.1 - Fully working prototype.
