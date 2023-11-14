package in_memory

import "github.com/wissance/Ferrum/internal/domain"

// ServerData is used in managers.FileDataManager
type ServerData struct {
	Realms  []domain.Realm
	Clients []domain.Client
	Users   []any
}
