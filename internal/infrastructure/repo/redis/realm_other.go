package redis

import (
	"encoding/json"

	"github.com/wissance/Ferrum/internal/domain"
	sf "github.com/wissance/stringFormatter"
)

// Creates a realm, if the realm has users and clients they will also be created.
// realmValue - json
func (mn *RedisRepo) CreateRealm(realmValue []byte) (*domain.Realm, error) {
	// TODO не забыть про аналог транзакции, что делать если ошибка
	var realm domain.Realm
	err := json.Unmarshal(realmValue, &realm)
	if err != nil {
		mn.logger.Error(sf.Format("An error occurred during Realm unmarshall"))
		return nil, err
	}

	// TODO выдавать ошибку если существуте такой client и user. И если есть такой realm
	if len(realm.Clients) != 0 {
		// TODO возможно нужно проверять, что есть какие-то поля у clients
		realmClients := make([]domain.ExtendedIdentifier, len(realm.Clients))
		for i, client := range realm.Clients {
			bytesClient, err := json.Marshal(client)
			if err != nil {
				return nil, err
			}
			clientKey := sf.Format(clientKeyTemplate, mn.namespace, client.Name)
			if err := setString(mn.redisClient, mn.ctx, mn.logger, Client, clientKey, string(bytesClient)); err != nil {
				return nil, err
			}
			realmClients[i] = domain.ExtendedIdentifier{
				ID:   client.ID,
				Name: client.Name,
			}
		}
		if err := mn.createRealmClients(realm.Name, realmClients, true); err != nil {
			return nil, err
		}
	}

	if len(realm.Users) != 0 {
		realmUsers := make([]domain.ExtendedIdentifier, len(realm.Users))
		for i, user := range realm.Users {
			bytesUser, err := json.Marshal(user)
			if err != nil {
				return nil, err
			}
			user := domain.NewUser(user)
			userName := user.GetUsername()
			userId := user.GetId()

			userKey := sf.Format(userKeyTemplate, mn.namespace, userName)
			if err := setString(mn.redisClient, mn.ctx, mn.logger, User, userKey, string(bytesUser)); err != nil {
				return nil, err
			}
			realmUsers[i] = domain.ExtendedIdentifier{
				ID:   userId,
				Name: userName,
			}
		}
		if err := mn.createRealmUsers(realm.Name, realmUsers, true); err != nil {
			return nil, err
		}
	}

	shortRealm := domain.Realm{
		Name:                   realm.Name,
		Clients:                []domain.Client{},
		Users:                  []any{},
		TokenExpiration:        realm.TokenExpiration,
		RefreshTokenExpiration: realm.RefreshTokenExpiration,
	}
	jsonShortRealm, err := json.Marshal(shortRealm)
	if err != nil {
		return nil, err
	}
	realmKey := sf.Format(realmKeyTemplate, mn.namespace, realm.Name)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, Realm, realmKey, string(jsonShortRealm)); err != nil {
		return nil, err
	}
	// TODO надо подумать над возвратом. есть ли смысл возвращать users, если при get у нас users не возвращаются
	return &realm, nil // TODO нет смысла возвращать реалм без client и user. Нужно наверное делать глубокую копию
}

// Removes realmClietns and realmUsers. Does not delete clients and users
func (mn *RedisRepo) DeleteRealm(realmName string) error {
	// TODO добавить ошибку если такого realm нет
	// TODO добавить транзитивность.
	realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realmName)
	redisIntCmd := mn.redisClient.Del(mn.ctx, realmClientsKey)
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}

	realmUsersKey := sf.Format(realmUsersKeyTemplate, mn.namespace, realmName)
	redisIntCmd = mn.redisClient.Del(mn.ctx, realmUsersKey)
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}

	realmKey := sf.Format(realmKeyTemplate, mn.namespace, realmName)
	redisIntCmd = mn.redisClient.Del(mn.ctx, realmKey)
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}

	return nil
}

func (mn *RedisRepo) UpdateRealm(realmName string, realmValue []byte) (*domain.Realm, error) {
	oldRealm, err := mn.GetRealm(realmName)
	if err != nil {
		return nil, err
	}
	var newRealm domain.Realm
	if err := json.Unmarshal(realmValue, &newRealm); err != nil {
		return nil, err
	}
	if oldRealm.Name != newRealm.Name {
		// TODO каскадно обновлять информацию у всех клиентов и user у realm. И удалить сам realm
	}

	shortRealm := domain.Realm{
		Name:                   newRealm.Name,
		Clients:                []domain.Client{},
		Users:                  []any{},
		TokenExpiration:        newRealm.TokenExpiration,
		RefreshTokenExpiration: newRealm.RefreshTokenExpiration,
	}
	jsonShortRealm, err := json.Marshal(shortRealm)
	if err != nil {
		return nil, err
	}
	realmKey := sf.Format(realmKeyTemplate, mn.namespace, shortRealm.Name)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, Realm, realmKey, string(jsonShortRealm)); err != nil {
		return nil, err
	}
	return &shortRealm, nil
}
