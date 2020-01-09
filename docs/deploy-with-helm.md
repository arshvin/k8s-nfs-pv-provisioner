# Deployment with Helm chart
In order to deploy the provisioner by Helm chart in vagrant machine it is needed to
1. install `helm3` command if any with help the [page](https://helm.sh/docs/intro/install/) (maybe later it will be in provisioning machine stage)
2. execute the commands like:
    ```bash
    [vagrant@microk8s ~]$ cd /vagrant/deploy/chart
    [vagrant@microk8s ~]$ kubectl create namespace pv-provisioner
    [vagrant@microk8s ~]$ helm upgrade -i pv-provisioner --namespace pv-provisioner -f values-overrides/values.yaml ./
    ```
3. check the result:
    ```bash
    [vagrant] ~ > kubectl -n pv-provisioner get pods
    NAME                              READY   STATUS    RESTARTS   AGE
    pv-provisioner-7c565c45db-dwvns   1/1     Running   0          9s

    [vagrant] ~ > kubectl -n pv-provisioner logs persistent-volume-provisioner-776b74b46f-d89zk -f
    I0315 13:06:43.427158       1 root.go:84] Trying to use in-cluster config
    I0315 13:06:43.463220       1 common.go:39] Starting controller: PersistentVolume
    I0315 13:06:43.463486       1 common.go:39] Starting controller: PersistentVolumeClaim
    ...
    ```
