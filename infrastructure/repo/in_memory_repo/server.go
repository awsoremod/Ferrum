package data

// ServerData is used in managers.FileDataManager
type ServerData struct {
	Realms  []Realm
	Clients []Client
	Users   []any
}
