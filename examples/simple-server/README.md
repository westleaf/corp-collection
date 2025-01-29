# Simple-server
This simple-server is meant to temporarily serve a folder inside a kuberentes cluster.
It was initially made as an alternative to `python3 -m http.server 80`.
Do not use in production.

## Example manifest
```YAML
apiVersion: v1
kind: Pod
metadata:
  name: simple-server-pod
  labels:
    app: simple-server
spec:
  containers:
  - name: fileserver-helper
    image: ghcr.io/westleaf/simple-server:latest
    imagePullPolicy: Always
    command: [ '/main', '--port', '8100', '--path', '/build' ]
    ports:
    - containerPort: 8100
      name: default-port
    volumeMounts:
    - name: files-to-serve-claim
      mountPath: /build
    resources:
      requests:
        memory: 256Mi
      limits:
        memory: 256Mi
  volumes:
  - name: files-to-serve-claim
    persistentVolumeClaim:
      claimName: files-to-serve
  imagePullSecrets:
    - name: ghcr-secret
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: files-to-serve
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: v1
kind: Service
metadata:
  name: simple-server-service
spec:
  selector:
    app: simple-server
  ports:
    - name: simple-server
      protocol: TCP
      port: 8080
      targetPort: default-port
```
