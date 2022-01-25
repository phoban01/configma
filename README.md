# ConfigMa 

***Stable Kubernetes ConfigMaps in an unstable world***

ConfigMa enables you to present a persistent ConfigMap to your application using the data from the latest instance of an immutable ConfigMap.

For example, let's imagine your pipeline produces ConfigMaps with the following naming convention:

```shell
$ kubectl get cm
NAME                              DATA   AGE
network-conf-be1c2b               1      10m
network-conf-89fea1               1      12m
network-conf-c5b12f               1      15m
```

This presents a difficulty if you wish to use this data in a Pod:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: network-pod
spec:
  containers:
    - name: net-container
      image: network-image
      volumeMounts:
      - name: config-volume
        mountPath: /etc/config
  volumes:
    - name: config-volume
      configMap:
        name: network-conf-be1c2b
```

Although our application can read the data as it changes in the ConfigMap, we need to find a way to update the ConfigMap reference in the pod each time a new version of the ConfigMap is generated with the latest network configuration data.

This is the problem ConfigMa aims to solve. Using the `ConfigMatcher` custom resource we now do the following:

```yaml
apiVersion: util.phoban.io/v1alpha1
kind: ConfigMatch
metadata:
  name: configmatch-sample
spec:
  sourceRef:
    kind: ConfigMap
    pattern: ^network-conf-*$
    matchGroup: config
  target:
    kind: ConfigMap
    name: network-conf
    namespace: default
```
We must also add the following label to our ConfigMaps: `configma.io/group: config`.

With this in place ConfigMa will now generate a ConfigMap named `network-conf` in the `default` namespace and keep it's data in sync with the latest version of the ConfigMap matching the regex pattern `.spec.sourceRef.pattern` with the matching group.

Your workload can now use this stable reference:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: network-pod
spec:
  containers:
    - name: net-container
      image: network-image
      volumeMounts:
      - name: config-volume
        mountPath: /etc/config
  volumes:
    - name: config-volume
      configMap:
        name: network-conf
```

Inspired by https://github.com/gopaddle-io/configurator
