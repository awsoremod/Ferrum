package redis_repo

import (
	"encoding/json"
	"fmt"

	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/infrastructure/repo/errors_repo"

	sf "github.com/wissance/stringFormatter"
)

func (mn *RedisDataManager) CreateClient(clientValue []byte) (*data.Client, error) {
	// TODO не забыть про транзакции

	// TODO возможно нужно проверять, что есть какие-то поля у clients
	// должны быть {"id": "", "name": "", "type": "", "auth": {"type": 1, "value": ""}}
	//

	// TODO выдавать ошибку если такой client есть

	var client data.Client
	err := json.Unmarshal(clientValue, &client)
	if err != nil {
		mn.logger.Error(sf.Format("An error occurred during Client unmarshall")) // todo при выходе мб эту ошибку обработать и сказать, что value не правильный
		return nil, err
	}
	clientKey := sf.Format(clientKeyTemplate, mn.namespace, client.Name)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, Client, clientKey, string(clientValue)); err != nil {
		return nil, err
	}
	return &client, nil
}

func (mn *RedisDataManager) AddClientToRealm(realmName string, clientName string) error {
	// TODO выдавать ошибку если такой клиент есть в realm
	client, err := mn.GetClient(clientName)
	if err != nil {
		return err
	}
	realmClient := data.ExtendedIdentifier{
		ID:   client.ID,
		Name: client.Name,
	}
	sliceRealmClient := []data.ExtendedIdentifier{realmClient}
	if err := mn.createRealmClients(realmName, sliceRealmClient, false); err != nil {
		return err
	}
	return nil
}

func (mn *RedisDataManager) createRealmClients(realmName string, realmClients []data.ExtendedIdentifier, isAllPreDelete bool) error {
	bytesRealmClients, err := json.Marshal(realmClients)
	if err != nil {
		return err
	}
	realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realmName)
	if isAllPreDelete {
		redisIntCmd := mn.redisClient.Del(mn.ctx, realmClientsKey)
		if redisIntCmd.Err() != nil {
			return redisIntCmd.Err()
		}
	}
	redisIntCmd := mn.redisClient.RPush(mn.ctx, realmClientsKey, string(bytesRealmClients))
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}
	return nil
}

func (mn *RedisDataManager) DeleteClient(clientName string) error {
	// TODO добавить удаление во всех realm этого клиента
	clientKey := sf.Format(clientKeyTemplate, mn.namespace, clientName)
	redisIntCmd := mn.redisClient.Del(mn.ctx, clientKey)
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}
	return nil
}

// Deletes client from realmClients, does not delete client
func (mn *RedisDataManager) DeleteRealmClient(realmName string, clientName string) error {
	// todo много лишнего. для удаления склиента. происходит получение клиентов. нахождение клиента, удаление его из массива.
	// удаление всех клиентов из редис. и добавление нового массива клиентов в редис

	// TODO выдавать ошибку если такого слиента нет в realm
	realmClients := mn.getRealmClients(realmName)
	if len(realmClients) == 0 {
		return fmt.Errorf("getRealmClients failed: %w", errors_repo.ErrZeroLength)
	}
	isHasClient := false
	for i := range realmClients {
		if realmClients[i].Name == clientName {
			realmClients[i] = realmClients[len(realmClients)-1]
			realmClients = realmClients[:len(realmClients)-1]
			isHasClient = true
		}
	}
	if !isHasClient {
		return fmt.Errorf("realm \"%s\" doesn't have client \"%s\" in Redis: %w", realmName, clientName, errors_repo.ErrNotFound)
	}
	if err := mn.createRealmClients(realmName, realmClients, true); err != nil {
		return err
	}
	return nil
}

func (mn *RedisDataManager) UpdateClient(clientName string, clientValue []byte) (*data.Client, error) {
	oldClient, err := mn.GetClient(clientName)
	if err != nil {
		return nil, err
	}
	var newClient data.Client
	if err := json.Unmarshal(clientValue, &newClient); err != nil {
		return nil, err
	}
	if newClient.ID != oldClient.ID || newClient.Name != oldClient.Name {
		// TODO каскадно обновлять информацию во всех realm где был этот клиент. И удалить сам клиент
	}
	clientKey := sf.Format(clientKeyTemplate, mn.namespace, newClient.Name)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, Client, clientKey, string(clientValue)); err != nil {
		return nil, err
	}
	return &newClient, nil
}
