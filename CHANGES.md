# Change list
* 0.2.6 - Bug fixes
* 0.2.0 - Added:
  1. ability to pass more than one storage class name through CLI flag as comma separated list of names.
  2. processing of `ReclaimPolicy` of the storage class. Earlier the "Delete" reclaim policy was assigned to PV as hard coded value
  2. processing of `pv-provisioner/reclaim-policy` on PVC.
* 0.1.3 - Fully reworked algorithm of the provisioner to use Working queues and Informers
* 0.1.2 - Added support of `storage.asset/owner-uid` and `storage.asset/owner-gid` annotations for PVC.
* 0.1.1 - Fully working prototype.
