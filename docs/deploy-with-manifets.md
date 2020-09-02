# Deployment with manifest

In order to deploy the provisioner only using a manifests the command is used:
1. for install
    ```bash
    [vagrant] ~ > kubectl apply -f /vagrant/deploy/manifests/
    namespace/pv-provisioner created
    deployment.apps/persistent-volume-provisioner created
    storageclass.storage.k8s.io/nfs-storage created
    ```
2. for checking the result:

    ```bash
    [vagrant] ~ > kubectl -n pv-provisioner get pods
    NAME                                             READY   STATUS    RESTARTS   AGE
    persistent-volume-provisioner-776b74b46f-d89zk   1/1     Running   0          2m22s

    [vagrant] ~ > kubectl -n pv-provisioner logs persistent-volume-provisioner-776b74b46f-d89zk -f
    I0315 11:53:24.121762       1 root.go:84] Trying to use in-cluster config
    I0315 11:53:24.131468       1 common.go:39] Starting controller: PersistentVolume
    ...
    ```
