# NFS persistent volume provisioner
## Purpose
This stuff is intended to automate some routine activities during K8S usage like:
* addition/provisioning persistent volumes (hereafter _PV_) and binding them to corresponding persistent volume claims (_PVC_).
* removal/deprovisioning released _PV_ when owning _PVC_ had been removed.

The provisioner works similar as the _hostPath_ provisioner for Minikube and particularly it may use _hostPath_ volume type for creating new PV and bind to directory on host OS. However the provisioner may use _nfs_ type of PV. Therefore it makes a sense only for working with shared file systems like NFS where each of cluster nodes simultaneously can see the storage's changes.

## Related links
[Technical details about how it works](./docs/how-it-works.md).

[How to launch it for testing in your machine](./docs/getting-started.md).

Deployment notes for [Helm chart](./docs/deploy-with-helm.md) or [manifests](./docs/deploy-with-manifets.md).

[License](./LICENSE).

[Change](./CHANGES.md) list.
