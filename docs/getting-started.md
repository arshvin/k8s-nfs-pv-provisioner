# How to try and use it
In the root of the project there is the `Vagrantfile` which is used for provisioning [Vagrant](https://www.vagrantup.com/) machine and installing [microk8s](https://microk8s.io/). The Vagrant and [VirtualBox](https://www.virtualbox.org/) are required. To start machine (_VM_) it's needed to invoke from host shell:

```bash
vagrant up
```
The VM should be provisioned without any error in order to be ready to use. To check that VM is ready the couple of command will need to be invoked from the shell if provisioning process finished:

```bash
vagrant ssh
kubectl get node
docker ps
```

The success output will look like this:

```bash
[vagrant@microk8s ~]$ kubectl get node
NAME       STATUS   ROLES    AGE   VERSION
microk8s   Ready    <none>   14m   v1.15.6

[vagrant@microk8s ~]$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```
For PV it's needed to create directory `/mnt/nfs/`:
```bash
[vagrant@microk8s ~]$ sudo mkdir /mnt/nfs/
```
There are some manifests the `/vagrant/test/test_stuff/` directory to configure the test cluster and start to working on. To apply them from vagrant machine run the command:
```bash
[vagrant@microk8s ~]$ kubectl apply -f /vagrant/test/test_stuff/
namespace/test-01 created
storageclass.storage.k8s.io/nfs-storage created
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
```bash
[vagrant@microk8s ~]$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                               STORAGECLASS        REASON   AGE
pv01                                       5Gi        RWX            Retain           Available                                                                    13s
pv02                                       5Gi        RWO            Retain           Available                                                                    13s
pv03                                       5Gi        RWO            Delete           Available                                                                    13s
pv04                                       5Gi        RWO            Delete           Available                                                                    13s
pv05                                       5Gi        RWO            Delete           Available                                                                    13s
pvc-8bbada9a-5f13-478d-b9b0-aa9f60a52aef   20Gi       RWX            Delete           Bound       container-registry/registry-claim   microk8s-hostpath            10m

[vagrant@microk8s ~]$ kubectl get pvc -A
NAMESPACE            NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS         AGE
container-registry   registry-claim   Bound     pvc-8bbada9a-5f13-478d-b9b0-aa9f60a52aef   20Gi       RWX            microk8s-hostpath    11m
test-01              pvc01            Pending                                                                        nfs-storage          59s
test-01              pvc02            Pending                                                                        nfs-storage          59s
test-01              pvc03            Pending                                                                        nfs-storage          59s
test-01              pvc04            Pending                                                                        some-storage-class   59s
test-01              pvc05            Pending                                                                        nfs-storage          59s
test-01              pvc06            Pending                                                                        nfs-storage          59s

[vagrant@microk8s ~]$ kubectl get deployment -n test-01
NAME                 READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment01   0/1     1            0           83s
```
The PV _pvc-8bbada9a-5f13-478d-b9b0-aa9f60a52aef_ is used for the container-registry (see [details](https://microk8s.io/docs/registry-images)) which was deployed during provisioning process of VM. And moreover container-registry must be behind the reverse proxy with TLS to avoid a complains of docker and containerd about unsecure registry. But another PVs that look like _pv0{1..5}_ were created by one of applied manifest just recently.
The PVCs of _nfs-storage_ class are in pending status. They will be bound once the provisioner will start. At the current moment the storage class _nfs-storage_ does not exists yet and it will be created later.
To check that's OK it's needed to launch a couple commands:
```bash
[vagrant@microk8s:/vagrant]$ kubectl get pod -A
NAMESPACE            NAME                                    READY   STATUS    RESTARTS   AGE
container-registry   registry-6c99589dc-rq9wd                1/1     Running   0          30m
default              nginx-9fd4747f6-xd488                   1/1     Running   0          30m
kube-system          coredns-f7867546d-xq89m                 1/1     Running   0          30m
kube-system          hostpath-provisioner-65cfd8595b-v9fr5   1/1     Running   0          30m
test-01              nginx-deployment01-779cdd6cf8-cm5jw     0/1     Pending   0          20m

[vagrant@microk8s:/vagrant]$ kubectl get services -A
NAMESPACE            NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                  AGE
container-registry   registry     NodePort    10.152.183.78    <none>        5000:32000/TCP           32m
default              kubernetes   ClusterIP   10.152.183.1     <none>        443/TCP                  32m
default              nginx        NodePort    10.152.183.242   <none>        443:32001/TCP            32m
kube-system          kube-dns     ClusterIP   10.152.183.10    <none>        53/UDP,53/TCP,9153/TCP   32m
```
Our check list is:
1. working and ready to server pods like:
  * __container-registry/registry-*__ - registry
  * __kube-system/coredns-*__ - DNS serivice of K8s
  * __default/nginx-*__ - reverse proxy with TLS

2. services with endpoints:
  * __container-registry/registry__
  * __kube-system/kube-dns__
  * __default/nginx__

To build the docker image which will be deployed into K8S, the command should be launched. Let's for example the tag will be __0.0.1__:
```bash
docker build /vagrant/ -t microk8s:32001/nfs-provisioner:0.0.1
...
Successfully tagged microk8s:32001/nfs-provisioner:0.0.1

[vagrant@microk8s ~]$ docker push microk8s:32001/nfs-provisioner:0.0.1
...
0.0.1: digest: sha256:ec658fa209c9d64a2d79b7397e9e2ddca5a9d7e5ff2b8c75ff877b744f9227e0 size: 740
```
To deploy the provisioner in the Microk8s and to watch the result are left a couple of commands. There are 2 ways to do it:
1. to use the manifests (read [here](./deploy-with-manifets.md))
2. to use the Helm chart (read [here](./deploy-with-helm.md)):

When the provisioner started being working for the cluster the final steps are to check that PVCs with actual storage class were bound with just provisioned PVs:
```bash
[vagrant@microk8s ~]$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                               STORAGECLASS        REASON   AGE
pv01                                       5Gi        RWX            Retain           Available                                                                    106m
pv02                                       5Gi        RWO            Retain           Available                                                                    106m
pv03                                       5Gi        RWO            Delete           Available                                                                    106m
pv04                                       5Gi        RWO            Delete           Available                                                                    106m
pv05                                       5Gi        RWO            Delete           Available                                                                    106m
pvc-09fd6420-a0a6-4186-9295-928de0d507fb   20Gi       RWX            Delete           Bound       container-registry/registry-claim   microk8s-hostpath            137m
test-01-pvc01-vol                          10Gi       RWO            Delete           Bound       test-01/pvc01                       nfs-storage                  28m
test-01-pvc02-vol                          10Gi       RWX            Delete           Bound       test-01/pvc02                       nfs-storage                  28m
test-01-pvc03-vol                          5Gi        RWX            Delete           Bound       test-01/pvc03                       nfs-storage                  28m
test-01-pvc06-vol                          5Gi        RWX            Delete           Bound       test-01/pvc06                       nfs-storage                  28m

[vagrant@microk8s ~]$ kubectl get pvc -A
NAMESPACE            NAME             STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS         AGE
container-registry   registry-claim   Bound     pvc-8bbada9a-5f13-478d-b9b0-aa9f60a52aef   20Gi       RWX            microk8s-hostpath    28m
test-01              pvc01            Bound     test-01-pvc01-vol                          10Gi       RWO            nfs-storage          28m
test-01              pvc02            Bound     test-01-pvc02-vol                          10Gi       RWX            nfs-storage          28m
test-01              pvc03            Bound     test-01-pvc03-vol                          5Gi        RWX            nfs-storage          28m
test-01              pvc04            Pending                                                                        some-storage-class   28m
test-01              pvc05            Pending                                                                        nfs-storage          28m
test-01              pvc06            Bound     test-01-pvc06-vol                          5Gi        RWX            nfs-storage          28m
```
As it is expected the pvc05 was not provisioned because of the Selectors. Selectors are not supported yet.
