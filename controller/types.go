package controller

const (
	// VerloopManagedKey - The annotation which we look for in objects to sync
	VerloopManagedKey = "nsync.verloop.io/managed"
)

// ObjectType - DataType for K8s objects
type ObjectType uint8

const (
	_ ObjectType = iota
	// NAMESPACE - type to identify a NS
	NAMESPACE
	// CONFIGMAP - type to identify a ConfigMap
	CONFIGMAP
	// SECRET - type to identify a Secret
	SECRET
)

// ObjectName - Reverse map for Object type to their names
var ObjectName = map[ObjectType]string{
	NAMESPACE: "Namespace",
	CONFIGMAP: "Configmap",
	SECRET:    "Secret",
}

// Action - DataType to identify intended action on a k8s object
type Action uint8

const (
	// SKIP - Default action, equivalent to no annotation
	SKIP Action = iota
	// ENSURE - Ensures sync between current NS and any NS with the VerloopManagedKey annotation
	ENSURE
	// REMOVE - Removes Secret/ConfigMap from all NS
	REMOVE
)

// ActionName - Reverse map of actions for their string names.
var ActionName = map[Action]string{
	SKIP:   "Skip",
	ENSURE: "Ensure",
	REMOVE: "Remove",
}
