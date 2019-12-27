# NFS persistent volume provisioner
## Destination
This stuff is intended to automate some routine activities during K8S usage like:
* creating/provisioning persistent volumes (hereafter _PV_) and binding them to corresponding persistent volume claims (_PVC_).
* deleting/deprovisioning released _PV_ when owning _PVC_ had been removed.

The provisioner works similar as the _hostPath_ provisioner for Minikube and particularly it does use _hostPath_ volume type for creating new PV. Therefore it makes a sense only for working with shared file systems like NFS where each of cluster node simultaneously can see the storage's changes and the scheduled pod after the PV provisioning will be able to use it whenever node it would be started.

## How it works

The provisioner tries to follow to the [spec](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/storage/volume-provisioning.md). It does not have a deal with any remote storage APIs. All of it does for provisioning is 
* creates properly named directory under the `-inner-asset-root` which is mounted from the host OS into the provisioner's container
* creates _hostpath_ PV which points into the specified directory. 

### Details
#### Initialization stage

1. When the provisioner starts it reads from command line interface (_CLI_) flags in order to know what kind of unbound PVCs it should watch to create PV for them. The meaning flag for that is `-storage-class`.
2. From the CLI flags the provisioner discovers what directory it should use as root to create so called `storage asset` of PV. The meaning flag is `-inner-asset-root`
3. Another CLI flag says to provisioner whether it runs inside or outside of a k8s cluster. The meaning flag is `-kubectl-config`. 
4. Based on input argument's data the provisioner tries to connect to the cluster and get parameters of specified storage class. One of the storage class parameter is `outerAssetRoot` that is counterpart of specified value in CLI `-inner-asset-root` flag. These 2 parameters point at the same place on the shared file system. But the first one is used to create PV object and later will be used by k8s for mounting particular PV to a pod. The second one is used by only provisioner itself to create a storage asset by OS's syscall and therefore the second path must be mounted into provisioner's pod  (if it supposed to work inside the cluster). 
5. After that it gets started to cycle to looking for:
    * PVCs that need provisioned PVs. It is named `PV provisioning stage`
    * PVs that have been already released and may be deleted. It is named `PV deprovisioning stage`.

#### PV provisioning stage 
1. In order to determine a PVC that needs new created _PV_ the few conditions should be satisfied. The PVC:
    * should have "volume.beta.kubernetes.io/storage-provisioner" annotation with value equals to the name of actual provisioner. This value is set up by a controller, which is looking for it in the storage class
    * should have the same storage class as it was specified by `-storage-class` CLI flag
    * must not be bound to any _PV_
    * must not have any _Selectors_

    If any of the mentioned condition does not satisfied, the PVC is skipped and the provisioner is moving on to next one.
2. If the PVC is met to the conditions, the provisioner tries to create the _storage-asset_ (literally directory) under the path specified by `-inner-asset-root` flag. Naming convention of the new storage asset is "_namespace-of-pvc_-_name-of-pvc_-vol". The word "claim" if it is faced in the PVC name is removed. The ownership of the new created storage asset is assigned to user and group which are specified in parameters of the class storage `ownerNewAssetUid` and `ownerNewAssetGid` respectively.

    If the attempt of creating asset or setting of ownership are failed, the PVC is skipped and the provisioner is moving to the one.
3. If creating of the storage asset succeeded the provisioner tries to create a PV for corresponding PVC and bind them to each other. The naming convention for PV is the same as for storage asset: "_namespace-of-pvc_-_name-of-pvc_-vol".
    If the attempt is failed, the PVC is skipped and the provisioner is moving to next one.

#### PV deprovisioning stage
1. In order to determine PV that may be deleted the few conditions should be met. The PV
    * should have the annotation "pv.kubernetes.io/provisioned-by" value of which is the name of actual provisioner that specified in storage class.
    * should have `PersistentVolumeClaimPolicy` parameter which has value _Delete_

    If any of the mentioned conditions does not satisfied, the PV is skipped and the provisioner is moving to next one.
2. If the PV is met to the conditions, the provisioner tries to delete storage asset that can be deduced by parameters compilation like `PersistentVolumeSource.HostPath.Path` of the PV and `-inner-asset-root` of CLI flags of provisioner.

    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.

3. If storage asset deleting is succeeded the provisioner tries to delete the PV.
    
    If the attempt is failed, the PV is skipped and the provisioner is moving to next one.

## How to try and use it
In the root of the project there is the `Vagrantfile` which is used for provisioning [Vagrant](https://www.vagrantup.com/) machine and installing [microk8s](https://microk8s.io/). The Vagrant and [VirtualBox](https://www.virtualbox.org/) are required. To start machine it's needed to invoke from host shell:

```
vagrant up
```
The machine should be provisioned without any error in order to be ready to use. To check that machine is ready the couple of command will need to be invoked after provisioning from the shell:
```
vagrant ssh
kubectl get node
docker ps
```
The success output will look like this:
```
[vagrant@microk8s ~]$ kubectl get node
NAME       STATUS   ROLES    AGE   VERSION
microk8s   Ready    <none>   14m   v1.15.6

[vagrant@microk8s ~]$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```
For PV it's needed to create directory `/mnt/nfs/`:
```
[vagrant@microk8s ~]$ sudo mkdir /mnt/nfs/
```
There are some manifests the `/vagrant/test_k8s/` directory to configure the test cluster and start to working on. To apply them from vagrant machine run the command:
```
[vagrant@microk8s ~]$ kubectl apply -f /vagrant/test_k8s/
namespace/test-01 created
storageclass.storage.k8s.io/nfs-storage-class created
storageclass.storage.k8s.io/some-storage-class created
persistentvolume/pv01 created
persistentvolume/pv02 created
persistentvolume/pv03 created
persistentvolume/pv04 created
persistentvolume/pv05 created
persistentvolumeclaim/pvc01 created
persistentvolumeclaim/pvc03 created
persistentvolumeclaim/pvc04 created
persistentvolumeclaim/pvc05 created
deployment.apps/nginx-deployment01 created
persistentvolumeclaim/pvc02 created
```
And checking the result:
```
[vagrant@microk8s ~]$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                               STORAGECLASS        REASON   AGE
pv01                                       5Gi        RWX            Retain           Available                                                                    10s
pv02                                       5Gi        RWO            Retain           Available                                                                    10s
pv03                                       5Gi        RWO            Delete           Available                                                                    10s
pv04                                       5Gi        RWO            Delete           Available                                                                    10s
pv05                                       5Gi        RWO            Delete           Available                                                                    10s

[vagrant@microk8s ~]$ kubectl get pvc -A
NAMESPACE            NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS         AGE
microk8s-hostpath    73m
test-01              pvc01            Pending                                                                        nfs-storage-class    41m
test-01              pvc02            Pending                                                                        nfs-storage-class    41m
test-01              pvc03            Pending                                                                        nfs-storage-class    41m
test-01              pvc04            Pending                                                                        some-storage-class   41m
test-01              pvc05            Pending                                                                        nfs-storage-class    41m

[vagrant@microk8s ~]$ kubectl get deployment -n test-01
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment01   0/1     1            0           42m
```
The PVCs of nfs-storage-class are in pending status. They will be bound once the provisioner will start. 
In order to simplify working with docker images in microk8s (see [details](https://microk8s.io/docs/registry-images)) it's needed to deploy registry addon in it and add the registry as insecured:
```
[vagrant@microk8s ~]$ microk8s.enable registry
[vagrant@microk8s ~]$ sudo sh -c 'cat << EOF > /etc/docker/daemon.json
{
  "insecure-registries" : ["localhost:32000"]
}
EOF
'
[vagrant@microk8s ~]$ sudo systemctl restart docker
```
And it needs to add local repository as unsecured for `containerd` too (see [here](https://microk8s.io/docs/registry-private)):
```
[vagrant@microk8s ~]$ sudo sed -i /var/snap/microk8s/current/args/containerd-template.toml -e 's/local.insecure-registry.io/microks8:32000/'
[vagrant@microk8s ~]$ sudo sed -i /var/snap/microk8s/current/args/containerd-template.toml -e 's|http://localhost:32000|http://microks8:32000|'
[vagrant@microk8s ~]$ microk8s.stop
[vagrant@microk8s ~]$ microk8s.start
``` 
To build the docker image the command should be launched:
```
docker build /vagrant/ -t microk8s:32000/nfs-provisioner:0.0.1
...
Successfully tagged microk8s:32000/nfs-provisioner:0.0.1

[vagrant@microk8s ~]$ docker push microk8s:32000/nfs-provisioner:0.0.1
...
0.0.1: digest: sha256:ec658fa209c9d64a2d79b7397e9e2ddca5a9d7e5ff2b8c75ff877b744f9227e0 size: 740
```
To deploy the provisioner in the Microk8s and to watch the result are left a couple of commands:
```
vagrant@microk8s ~]$ sudo ctr image pull --plain-http microk8s:32000/nfs-provisioner:0.0.1
microk8s:32000/nfs-provisioner:0.0.1:                                                 resolved       |++++++++++++++++++++++++++++++++++++++| 
manifest-sha256:ec658fa209c9d64a2d79b7397e9e2ddca5a9d7e5ff2b8c75ff877b744f9227e0:     done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:30ffbd8805e1f00a01d04b1950c247319ff04eb617a83e5de9d9fc27c71178b0:        done           |++++++++++++++++++++++++++++++++++++++| 
config-sha256:49e1b2cbe4ad1bf10ad0060b7a5069d8dd48530aa773be16b1ecb00db0a5d6d4:       done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:9d48c3bd43c520dc2784e868a780e976b207cbf493eaff8c6596eb871cbd9609:        done           |++++++++++++++++++++++++++++++++++++++| 
elapsed: 0.2 s                                                                    total:  6.9 Mi (33.0 MiB/s)                                      
unpacking linux/amd64 sha256:ec658fa209c9d64a2d79b7397e9e2ddca5a9d7e5ff2b8c75ff877b744f9227e0...
done

[vagrant@microk8s ~]$ kubectl apply -f /vagrant/deploy/deployment.yml 
namespace/pv-provisioner unchanged
serviceaccount/pv-provisioner unchanged
clusterrole.rbac.authorization.k8s.io/pv-provisioner unchanged
clusterrolebinding.rbac.authorization.k8s.io/pv-provisioner unchanged
deployment.apps/persistent-volume-provisioner configured
The StorageClass "nfs-storage-class" is invalid: parameters: Forbidden: updates to parameters are forbidden. #It's OK!

[vagrant@microk8s ~]$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                               STORAGECLASS        REASON   AGE
pv01                                       5Gi        RWX            Retain           Available                                                                    106m
pv02                                       5Gi        RWO            Retain           Available                                                                    106m
pv03                                       5Gi        RWO            Delete           Available                                                                    106m
pv04                                       5Gi        RWO            Delete           Available                                                                    106m
pv05                                       5Gi        RWO            Delete           Available                                                                    106m
pvc-09fd6420-a0a6-4186-9295-928de0d507fb   20Gi       RWX            Delete           Bound       container-registry/registry-claim   microk8s-hostpath            137m
test-01-pvc01-vol                          10Gi       RWO            Delete           Bound       test-01/pvc01                       nfs-storage-class            2m8s
test-01-pvc02-vol                          8Gi        RWO            Delete           Bound       test-01/pvc02                       nfs-storage-class            2m8s
test-01-pvc03-vol                          5Gi        RWX            Delete           Bound       test-01/pvc03                       nfs-storage-class            2m8s
 
[vagrant@microk8s ~]$ kubectl get pvc -A
NAMESPACE            NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS         AGE
container-registry   registry-claim   Bound     pvc-09fd6420-a0a6-4186-9295-928de0d507fb   20Gi       RWX            microk8s-hostpath    163m
test-01              pvc01            Bound     test-01-pvc01-vol                          10Gi       RWO            nfs-storage-class    131m
test-01              pvc02            Bound     test-01-pvc02-vol                          8Gi        RWO            nfs-storage-class    131m
test-01              pvc03            Bound     test-01-pvc03-vol                          5Gi        RWX            nfs-storage-class    131m
test-01              pvc04            Pending                                                                        some-storage-class   131m
test-01              pvc05            Pending                                                                        nfs-storage-class    131m
``` 
pvc05 is expected was not provisioned because it has the Selectors.

The command `ctr image pull` was used because of the [bug/feature](https://github.com/containerd/cri/issues/1201) of Containerd cri.

## Helm chart
To deploy it by Helm chart it is needed to
1. install `helm3` command if any with help the [page](https://helm.sh/docs/intro/install/)
2. execute the command like:
```
[vagrant@microk8s ~]$ kubectl create namespace some-namespace
[vagrant@microk8s ~]$ helm install provisioner -f /vagrant/chart/values.yaml  /vagrant/chart/ --namespace some-namespace
``` 