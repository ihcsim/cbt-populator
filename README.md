# CBT Populator

![pipeline](https://github.com/ihcsim/cbt-populator/actions/workflows/pipeline.yaml/badge.svg)

CBT populator is a spike effort to build the CSI CBT service using a [volume
populator][1]. The full description of the CSI CBT service is described under
this [KEP][2].

## Deployment

### Prerequisites

In order to perform changed block tracking, the CBT populator uses the CSI 
`VolumeSnapshot` APIs to snapshot PVCs. Deploy the CSI `external-snapshotter` 
service following the instructions in its [documentation][4].

The example application depends on the [CSI hostpath driver][5] to provision
its PVC. Follow its [documentation][6] for installation.

### CBT Service

Deploy the CRD and controller's RBAC:

```sh
kubectl apply -f yaml/crd.yaml

kubectl auth reconcile -f yaml/rbac.yaml

kubectl create ns cbt-populator
```

Build and deploy the CBT controller and populator to a K8s cluster:

```sh
make apply KO_DOCKER_REPO=<your_image_registry>
```

Check the rollout status of the CBT controller:

```sh
kubectl -n cbt-populator rollout status deploy/cbt-populator 
```

Confirm that it is successfully rolled out:

```sh
deployment "cbt-populator" successfully rolled out
```

By default, the CBT controller runs in the `cbt-populator` namespace:

```sh
kubectl -n cbt-populator get po
```

Confirm that the controller pod is running:

```sh
NAME                             READY   STATUS    RESTARTS   AGE                              │➜  cbt-populator git:(main) ✗                                                           (
cbt-populator-59764d965b-8pd85   1/1     Running   0          17h
```

### Example Application

Deploy the CSI hostpath driver's storage class:

```sh
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/csi-driver-host-path/ed64506237332cd5c145a9e4cfec32dfae7a97b8/examples/csi-storageclass.yaml

kubectl get sc csi-hostpath-sc                                                                                                        
```

Confirm that the `csi-hostpath-sc` storage class is ready:

```sh
NAME              PROVISIONER           RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
csi-hostpath-sc   hostpath.csi.k8s.io   Delete          Immediate           true                   20s
```

Deploy the `src-pod` pod and its `src-pvc` PVC to the `default` namespace:

```sh
kubectl apply -f yaml/app/pod-pvc.yaml
```

This pod attaches the PVC at its `/data` mount point and prints the current 
timestamp to the `/data/date.txt` file every second:

```sh
kubectl exec src-pod -- /bin/sh -c "tail -f /data/date.txt"
```

Take two snapshots of the PVC, ten seconds apart:

```sh
kubectl apply -f ./yaml/app/volumesnapshot-from.yaml

sleep 10

kubectl apply -f ./yaml/app/volumesnapshot-to.yaml
```

These will create 2 `VolumeSnapshot` resources in the same namespace as the 
application:

```sh
kubectl get vs,vsc                                                                                                                             
```

Confirm that the volume snapshots and volume snapshot contents are ready:

```sh
NAME                                                 READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
volumesnapshot.snapshot.storage.k8s.io/cbt-vs-from   true         src-pvc                             10Mi          csi-hostpath-snapclass   snapcontent-f37fade4-70d8-47db-87d6-89e1f8bc83df   85s            85s
volumesnapshot.snapshot.storage.k8s.io/cbt-vs-to     true         src-pvc                             10Mi          csi-hostpath-snapclass   snapcontent-975a0133-e606-42f3-aba2-ab0702d214f4   46s            46s

NAME                                                                                             READYTOUSE   RESTORESIZE   DELETIONPOLICY   DRIVER                VOLUMESNAPSHOTCLASS      VOLUMESNAPSHOT   VOLUMESNAPSHOTNAMESPACE   AGE
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-f37fade4-70d8-47db-87d6-89e1f8bc83df   true         10485760      Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   cbt-vs-from      default                   85s
volumesnapshotcontent.snapshot.storage.k8s.io/snapcontent-975a0133-e606-42f3-aba2-ab0702d214f4   true         10485760      Delete           hostpath.csi.k8s.io   csi-hostpath-snapclass   cbt-vs-to        default                   46s
```

Create the pod and PVC to store the CBT mock data:

```sh
kubectl apply -f ./yaml/app/cbt-pvc.yaml
```

```sh 
kubectl get po,pvc,cbt 
```

```sh
NAME                       READY   STATUS    RESTARTS   AGE
pod/src-pod                1/1     Running   0          4m33s
pod/cbt-pod                1/1     Running   0          110s

NAME                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/src-pvc   Bound    pvc-eea33546-930c-4129-9b43-45d4fe2ca854   10Mi       RWO            csi-hostpath-sc   4m33s
persistentvolumeclaim/cbt-pvc   Bound    pvc-892646d4-3a13-4621-aead-a2e19c35d98f   10Mi       RWO            csi-hostpath-sc   110s

NAME                                           AGE
changedblockrange.cbt.storage.k8s.io/cbt-cbr   110s
```

Observe that the populator pod and prime PVC are created in the `cbt-populator` 
namespace:
```sh
$ kubectl -n cbt-populator get po,pvc                                                                                                             
```

```sh
NAME                                                READY   STATUS              RESTARTS   AGE
pod/cbt-populator-5d8b5b769f-ddz42                  1/1     Running             0          3m31s
pod/populate-8202ccd0-24d9-40e1-ab5d-1d7dadb09b21   0/1     ContainerCreating   0          2s

NAME                                                               STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/prime-8202ccd0-24d9-40e1-ab5d-1d7dadb09b21   Bound    pvc-56d814e9-1ede-4eca-9aa5-76d140ddd75a   10Mi       RWO            csi-hostpath-sc   28s
```

Confirm that the mock data is stored in the `cbt-pvc` volume:

```sh
kubectl exec cbt-pod -- ls /data/
```

There should be a data file in the `/data` folder:
```sh
cbt-1681146460
```

Use `cat` to view the content of the data file:
```sh
kubectl exec cbt-pod -- cat /data/cbt-1681146460
```

Expect to see a JSON CBT payload:

```json
{
  "ChangedBlocks": [
    {
      "BlockIndex": 0,
      "FirstBlockToken": "AAABAVahm9SO60Dyi0ORySzn2ZjGjW/KN3uygGlS0QOYWesbzBbDnX2dGpmC",
      "SecondBlockToken": "AAABAf8o0o6UFi1rDbSZGIRaCEdDyBu9TlvtCQxxoKV8qrUPQP7vcM6iWGSr"
    },
    {
      "BlockIndex": 6000,
      "FirstBlockToken": "AAABAbYSiZvJ0/R9tz8suI8dSzecLjN4kkazK8inFXVintPkdaVFLfCMQsKe",
      "SecondBlockToken": "AAABAZnqTdzFmKRpsaMAsDxviVqEI/3jJzI2crq2eFDCgHmyNf777elD9oVR"
    },
    {
      "BlockIndex": 6001,
      "FirstBlockToken": "AAABASBpSJ2UAD3PLxJnCt6zun4/T4sU25Bnb8jB5Q6FRXHFqAIAqE04hJoR"
    },
    {
      "BlockIndex": 6002,
      "FirstBlockToken": "AAABASqX4/NWjvNceoyMUljcRd0DnwbSwNnes1UkoP62CrQXvn47BY5435aw"
    },
    {
      "BlockIndex": 6003,
      "FirstBlockToken": "AAABASmJ0O5JxAOce25rF4P1sdRtyIDsX12tFEDunnePYUKOf4PBROuICb2A"
    },
  ],
  "ExpiryTime": 1576308931.973,
  "VolumeSize": 32212254720,
  "BlockSize": 524288,
  "NextToken": "AAADARqElNng/sV98CYk/bJDCXeLJmLJHnNSkHvLzVaO0zsPH/QM3Bi3zF//O6Mdi/BbJarBnp8h"
  }
```

## Development

The Makefile defines a number of targets for developing the CBT controller and
populator, based on [`ko`][3]. 

Install `ko` following the instructions described in its [documentation][3].

Run the test:

```sh
make test
```

Perform local build:

```sh
make build
```

Build and push the images:

```sh
make push KO_DOCKER_REPO=<your_image_registry>
```

To re-generate the API Go code:

```sh
make codegen
```

## License

See the [LICENSE](LICENSE) file.

[1]: https://kubernetes.io/blog/2022/05/16/volume-populators-beta/
[2]: https://github.com/kubernetes/enhancements/pull/3367
[3]: https://ko.build/install/
[4]: https://github.com/kubernetes-csi/external-snapshotter#usage
[5]: https://github.com/kubernetes-csi/csi-driver-host-path
[6]: https://github.com/kubernetes-csi/csi-driver-host-path/blob/ed64506237332cd5c145a9e4cfec32dfae7a97b8/docs/deploy-1.17-and-later.md
