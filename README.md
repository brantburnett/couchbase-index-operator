# Couchbase Index Controller

The Couchbase Index Controller manages Couchbase indices in Kubernetes declaratively, vastly
simplifying the management of indices as part of your deployment pipeline. Once installed,
indices may be managed simply using `kubectl apply`, templated using [Kustomize](https://kustomize.io/),
or using any other deployment pipeline that supports [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

The operator works within the [Operator Framework](https://operatorframework.io/), using a
reconcilation cycle to compare the declared indices to the actual indices present in Couchbase.
It then makes any changes required to align the infrastructure with the desired state. It
does so using [couchbase-index-manager](https://github.com/brantburnett/couchbase-index-manager)
under the hood.

> :information_source: The operator will not delete any indices it doesn't manage, so it plays well alongside other
> index management approaches. The only indices it will delete are indices that are removed from a
> CouchbaseIndexSet or when an entire CouchbaseIndexSet is deleted.

The operator pairs well with the [Couchbase Autonomous Operator](https://www.couchbase.com/products/cloud/kubernetes),
pulling configuration and authentication directly from the Couchbase-managed CRDs. However, it
may target any accessible Couchbase cluster.

## Installing

Example installations are available in the `docs/install` directory.

```sh
kubectl apply -f https://raw.githubusercontent.com/brantburnett/couchbase-index-operator/main/docs/install/v1.0.0.yaml
```

This will deploy the operator in the default namespace, and it will only operate on resources
within the default namespace. The example may be downloaded and the namespace changed to target
another namespace.

### Namespacing

By default, the operator will monitor all namespaces for CouchbaseIndexSet resources. This may
present security concerns or difficulties when testing upgrades of the operator.

It is recommended to deploy the operator to target a specific namespace using
`--watch-namespace=your-namespace` or the `WATCH_NAMESPACE` environment variable.

> :information_source: The example deployment fills the `WATCH_NAMESPACE` environment variable automatically with the
> namespace where the operator is deployed.

### Using a custom couchbase-index-manager image

By default, the latest released version of couchbase-index-manager is used (as of the time of the operator build).
To use an alternative version, supply it on the command line for the operator: `--cbim-image=btburnett3/couchbase-index-manager:1.0.1`
or via the `CBIM_IMAGE` environment variable.

## Deploying Indices

Indices are defined as a `CouchbaseIndexSet` resource in Kubernetes. Grouping multiple indices into
a set allows us to combine the index build operation for multiple indices into a single run in
Couchbase Server.

```yaml
apiVersion: couchbase.btburnett.com/v1beta1
kind: CouchbaseIndexSet
metadata:
  name: couchbaseindexset-sample
spec:
  cluster: 
    clusterRef:
      # Name of the CouchbaseCluster resource in Kubernetes created for the Couchbase Autonomous Operator
      # It must reside in the same namespace as the CouchbaseIndexSet
      name: cb-example 
      # Optional secret name, for use if you don't want to use the AdminSecretName from the CouchbaseCluster resource
      secretName: cb-example-auth
  # If using managed buckets within the CouchbaseCluster resource, this bucket name will be validated against the list of managed buckets
  bucketName: default
  indices:
  - name: example
    indexKey:
    - id
    condition: type = 'airline'
    replicas: 1
  - name: example2
    indexKey:
    - type
    - name
    - id
    partition:
      expressions:
      - meta().id
```

### Targeting an externally managed Couchbase cluster

To target an externally managed Couchbase cluster, use `manual` instead of `clusterRef`.
The referenced [Secret](https://kubernetes.io/docs/concepts/configuration/secret/) must exist
in the same namespace as the CouchbaseIndexSet, and contain two keys `username` and `password`.

```yaml
apiVersion: couchbase.btburnett.com/v1beta1
kind: CouchbaseIndexSet
metadata:
  name: couchbaseindexset-sample
spec:
  cluster: 
    manual:
      # Name of the CouchbaseCluster resource in Kubernetes
      connectionString: couchbase://cb-example
      # Name of the secret containing username and password keys
      # It must reside in the same namespace as the CouchbaseIndexSet
      secretName: cb-example-auth
  # Using manual the bucketName is not validated before the sync run
  bucketName: default
  indices:
  - name: example
    indexKey:
    - id
    condition: type = 'airline'
    replicas: 1
  - name: example2
    indexKey:
    - type
    - name
    - id
    partition:
      expressions:
      - meta().id
```

### Controlling run time

By default, sync jobs are given 5 minutes to complete, and will retry 2 additional times after a failure.
This is adequate in many cases, but may be insufficient for large indices. The `activeDeadlineSeconds` and
`backoffLimit` fields can adjust how long the created sync Job will run and how many times it will retry.

```yaml
apiVersion: couchbase.btburnett.com/v1beta1
kind: CouchbaseIndexSet
metadata:
  name: couchbaseindexset-sample
spec:
  cluster: 
    clusterRef:
      name: cb-example 
  bucketName: default
  indices:
  - name: example
    indexKey:
    - id
    condition: type = 'airline'
    replicas: 1
  activeDeadlineSeconds: 1200
  backoffLimit: 0
```

## Monitoring Status

The status of the indices may be monitored using `kubectl describe`. The `Ready` condition will be `True` if the indices have been fully built and are in sync. the `Syncing` condition indicates if a sync is currently in progress.

Note that `Ready` may be `True` when `Syncing` is also true. The indices are resynced regularly to ensure that indices remain in place, even if they are lost on the Couchbase side due to node failure or accidental deletion.

Events are also emitted when the sync starts or stops.

```sh
> kubectl describe couchbaseindexset couchbaseindexset-sample

Name:         couchbaseindexset-sample
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  couchbase.btburnett.com/v1beta1
Kind:         CouchbaseIndexSet
Metadata:
  Creation Timestamp:  2021-08-23T21:08:17Z
  Finalizers:
    couchbase.btburnett.com/indices
  Generation:  1
  Resource Version:  5111421
  UID:               9ccbd6ab-cb47-4a27-85fc-60bfc436dc20
Spec:
  Active Deadline Seconds:  600
  Backoff Limit:            2
  Bucket Name:              default
  Cluster:
    Cluster Ref:
      Name:  cb-example
  Indices:
    Index Key:
      type
      id
    Name:  example
    Index Key:
      type
      name
      id
    Name:  example2
    Partition:
      Expressions:
        meta().id
      Strategy:  Hash
Status:
  Conditions:
    Last Transition Time:  2021-08-23T21:08:25Z
    Message:               Indices are in sync
    Observed Generation:   1
    Reason:                InSync
    Status:                True
    Type:                  Ready
    Last Transition Time:  2021-08-23T21:08:25Z
    Message:               Sync not in progress
    Observed Generation:   1
    Reason:                NotSyncing
    Status:                False
    Type:                  Syncing
  Config Map Name:         couchbaseindexset-sample-indexspec
  Index Count:             2
  Indices:
    example
    example2
Events:
  Type    Reason        Age   From                            Message
  ----    ------        ----  ----                            -------
  Normal  SyncStarted   15s   couchbase-index-set-controller  Sync started
  Normal  SyncComplete  7s    couchbase-index-set-controller  Sync completed
```

### Failure Conditions

It is possible for an index sync to fail for a variety of reasons. Therefore, the pods which are responsible for performing the sync are left in place for 15 minutes. This provides the opportunity to use `kubectl logs` to extract failure logs.

## Development

Developing locally is best supported using Kubernetes deployed locally using
Docker Desktop and VSCode Remote Containers. Once the folder is opened as a
Remote Container in VSCode you will have a command prompt with all required
tooling preinstalled.

```sh
# Build the project
make

# Build CRD manifests
make manifests

# Build the Docker image
make docker-build

# Deploy manifests and the operator to your current Kubernetes cluster
# Note: assumes `make manifests` and `make docker-build` are already run
make deploy

# Deploy the sample
kubectl apply -k ./config/samples

# Remove everything from your current Kubernetes Cluster
make undeploy

# Run unit tests
make test
```
