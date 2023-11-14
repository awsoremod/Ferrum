package redis

import (
	"fmt"

	"github.com/wissance/Ferrum/internal/domain"
	sf "github.com/wissance/stringFormatter"
)

func (mn *RedisRepo) GetRealm(realmName string) (*domain.Realm, error) {
	realmKey := sf.Format(realmKeyTemplate, mn.namespace, realmName)
	realm := getObjectFromRedis[domain.Realm](mn.redisClient, mn.ctx, mn.logger, Realm, realmKey)
	if realm == nil {
		return nil, fmt.Errorf("not found") // TODO(sia)
	}
	return realm, nil
}

// GetRealm function for getting realm by name
/* This function constructs Redis key by pattern combines namespace and realm name (realmKeyTemplate). Unlike from FILE provider
 * Realm stored in Redis does not have Clients and Users inside Realm itself, these objects must be queried separately.
 * Parameters:
 *     - realmName name of a realm
 * Returns: realms or nil (if realm was not found)
 */
func (mn *RedisRepo) GetRealmWithClients(realmName string) (*domain.Realm, error) {
	realm, err := mn.GetRealm(realmName)
	if err != nil {
		return nil, err
	}

	// should get realms too
	// if realms were stored without clients (we expected so), get clients related to realm and assign here
	if len(realm.Clients) == 0 {
		clients, err := mn.GetClientsFromRealm(realmName)
		if err != nil {
			return nil, err
		}
		realm.Clients = clients
	}
	return realm, nil
}
