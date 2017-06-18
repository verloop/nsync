package controller

const (
	VERLOOP_MANAGED_KEY = "verloop.io/managed"
)

type ObjectType uint8

const (
	_ ObjectType = iota
	NAMESPACE
	CONFIGMAP
	SECRET
)

var ObjectName = map[ObjectType]string{
	NAMESPACE: "Namespace",
	CONFIGMAP: "Configmap",
	SECRET:    "Secret",
}

type Action uint8

const (
	SKIP Action = iota
	ENSURE
	REMOVE
)

var ActionName = map[Action]string{
	SKIP:   "Skip",
	ENSURE: "Ensure",
	REMOVE: "Remove",
}
