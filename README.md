# Clear Kubernetes Node Conditions

The clear-conditions script is designed to facilitate managing conditions tied to a node in a cluster. It specifically addresses the issue where conditions, once set, do not automatically cease even when there is no update for a substantial duration. This is especially helpful where changes in the infrastructure on the node lead to the persistence of unwanted conditions.

## How it Works
The script operates by resetting all conditions and maintains only four standard ones (Ready, MemoryPressure, DiskPressure and PIDPressure). If there are services in the cluster that impose their own non-standard conditions (for example, Node Problem Detector), they appear within a few minutes.

## Usage
1.	To clear the conditions for a single node: `clear-conditions --context <CLUSTER_NAME> <NODE_NAME>`

2.	Upon execution, the script will display the list of nodes, context name, kube-config and prompt for operation confirmation. To give confirmation in advance, use the --yes flag: `clear-conditions --yes --context <CLUSTER_NAME> <NODE_NAME>`

3.	To reset conditions for all nodes in cluster, use the --all parameter. However, note that in larger clusters, this script may not reset conditions for certain nodes due to other ongoing processes. The script will highlight nodes where the reset was not successful, and in such cases, running the script separately for each of these nodes is advised. Also you could try to use following one-liner as an alternative for --all parameter: `kubectl get nodes --context <CLUSTER_NAME> | awk 'NR>1 {print "clear-conditions --yes --context <CLUSTER_NAME> "$1""}' | bash`

4.	To specify a path to kube-config if the essential context resides in a separate config: `clear-conditions --context <CLUSTER_NAME> --kubeconfig <PATH> <NODE_NAME>`

5.	A conditional overwrite mode is available but typically not required. It allows users to replace the primary conditions with default values (Ready=True, MemoryPressure=False, DiskPressure=False, PIDPressure=False). While this can be helpful with certain condition-related issues, it isn't recommended due to potential temporary changes in the node status from NotReady to Ready. `clear-conditions --overwrite <NODE_NAME>`

## Links
- [Node Conditions in Kubernetes Docs](https://kubernetes.io/docs/reference/node/node-status/#condition)
- [Node Problem Detector](https://github.com/kubernetes/node-problem-detector)

