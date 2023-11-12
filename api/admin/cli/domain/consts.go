package domain

type ResourceType string

const (
	RealmResource  ResourceType = "realm"
	ClientResource              = "client"
	UserResource                = "user"
)

type OperationType string

const (
	GetOperation    OperationType = "get"
	CreateOperation               = "create"
	DeleteOperation               = "delete"
	UpdateOperation               = "update"
)
