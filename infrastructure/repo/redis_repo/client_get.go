package redis_repo

import (
	"fmt"

	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/infrastructure/repo/errors_repo"
	sf "github.com/wissance/stringFormatter"
)

// GetClientsFromRealm function for getting all realm clients
/* This function gets all realm client.
 * This function works in two steps:
 *     1. Get all data.ExtendedIdentifier pairs id-name
 *     2. Get all Client objects at once by key slices (every redis key for client combines from namespace, realm and client name)
 * Parameters:
 *    - realmName - name of the realm
 * Returns slice of Clients or nil
 */
func (mn *RedisDataManager) GetClientsFromRealm(realmName string) ([]data.Client, error) {
	realmClients, err := mn.getRealmClients(realmName)
	if err != nil {
		return err
	}

	clients := make([]data.Client, len(realmClients))
	for i, rc := range realmClients {
		// todo(UMV) get all them at once
		client, err := mn.GetClient(rc.Name)
		if err != nil {
			return nil, errors_repo.ErrNotFound
		}
		clients[i] = *client
	}

	return clients, nil
}

// GetClient function for get realm client by name
/* This function constructs Redis key by pattern combines namespace and realm name (realmClientsKeyTemplate)
 * Parameters:
 *     - realm - pointer to realm
 *     - name - client name
 * Returns: client or nil
 */
func (mn *RedisDataManager) GetClient(clientName string) (*data.Client, error) {
	clientKey := sf.Format(clientKeyTemplate, mn.namespace, clientName)
	client := getObjectFromRedis[data.Client](mn.redisClient, mn.ctx, mn.logger, Client, clientKey)
	if client == nil {
		return nil, fmt.Errorf("not found")
	}
	return client, nil
}

func (mn *RedisDataManager) GetClientFromRealm(realmName string, clientName string) (*data.Client, error) {
	realmClient, err := mn.getRealmClient(realmName, clientName)
	if err != nil {
		return nil, err
	}
	client, err := mn.GetClient(realmClient.Name)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (mn *RedisDataManager) getRealmClient(realmName string, clientName string) (*data.ExtendedIdentifier, error) {
	realmClients := mn.getRealmClients(realmName)
	if len(realmClients) == 0 {
		return nil, fmt.Errorf("Нет ни одного клиента в реалме") // TODO
	}

	realmHasClient := false
	var client data.ExtendedIdentifier
	for _, rc := range realmClients {
		if rc.Name == clientName {
			realmHasClient = true
			client = rc
			break
		}
	}
	if !realmHasClient {
		// TODO (sia) debug не показывается
		mn.logger.Debug(sf.Format("Realm: \"{0}\" doesn't have client: \"{1}\" in Redis", realmName, clientName))
		return nil, fmt.Errorf("not found") // TODO
	}

	return &client, nil
}

func (mn *RedisDataManager) getRealmClients(realmName string) ([]data.ExtendedIdentifier, error) {
	realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realmName)
	realmClients := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmClients, realmClientsKey)
	if len(realmClients) == 0 {
		mn.logger.Error(sf.Format("There are no clients for realm: \"{0}\" in Redis", realmName))
		return nil, fmt.Errorf("no clients in realm: %w", errors_repo.ErrZeroLength)
	}
	return realmClients, nil
}
