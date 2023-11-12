package redis_repo

import (
	"encoding/json"
	"fmt"

	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/infrastructure/repo/errors_repo"
	sf "github.com/wissance/stringFormatter"
)

func (mn *RedisDataManager) CreateUser(userValue []byte) (string, error) {
	// TODO выдавать ошибку если такой user есть
	var userFromJson any
	if err := json.Unmarshal(userValue, &userFromJson); err != nil {
		return "", err
	}
	// TODO проблема что происходит бесполезные машринг в следующей функции
	user := data.CreateUser(userFromJson)
	userName := user.GetUsername()
	userKey := sf.Format(userKeyTemplate, mn.namespace, userName)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, User, userKey, string(userValue)); err != nil {
		return "", err
	}
	// TODO возможно возвращать не имя. а объет распаршенный из value
	return userName, nil
}

func (mn *RedisDataManager) AddUserToRealm(realmName string, userName string) error {
	// Выдавать ошибку если такой user есть в realm
	user, err := mn.GetUser(userName)
	if err != nil {
		return err
	}
	userId := (*user).GetId()
	realmUser := data.ExtendedIdentifier{
		ID:   userId,
		Name: userName,
	}
	sliceRealmUser := []data.ExtendedIdentifier{realmUser}
	if err := mn.createRealmUsers(realmName, sliceRealmUser, false); err != nil {
		return err
	}
	return nil
}

func (mn *RedisDataManager) createRealmUsers(realmName string, realmUsers []data.ExtendedIdentifier, isAllPreDelete bool) error {
	bytesRealmUsers, err := json.Marshal(realmUsers)
	if err != nil {
		return err
	}
	realmUsersKey := sf.Format(realmUsersKeyTemplate, mn.namespace, realmName)
	if isAllPreDelete {
		redisIntCmd := mn.redisClient.Del(mn.ctx, realmUsersKey)
		if redisIntCmd.Err() != nil {
			return redisIntCmd.Err()
		}
	}
	redisIntCmd := mn.redisClient.RPush(mn.ctx, realmUsersKey, string(bytesRealmUsers))
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}
	return nil
}

func (mn *RedisDataManager) DeleteUser(userName string) error {
	// TODO добавить удаление во всех realm этого клиента
	userKey := sf.Format(userKeyTemplate, mn.namespace, userName)
	redisIntCmd := mn.redisClient.Del(mn.ctx, userKey)
	if redisIntCmd.Err() != nil {
		return redisIntCmd.Err()
	}
	return nil
}

// Deletes user from realmUsers, does not delete user
func (mn *RedisDataManager) DeleteRealmUser(realmName string, userName string) error {
	// TODO выдавать ошибку если такого клиента нет в realm
	realmUsers := mn.getRealmUsers(realmName)
	if len(realmUsers) == 0 {
		return fmt.Errorf("getRealmUsers failed: %w", errors_repo.ErrZeroLength)
	}
	isHasClient := false
	for i := range realmUsers {
		if realmUsers[i].Name == userName {
			realmUsers[i] = realmUsers[len(realmUsers)-1]
			realmUsers = realmUsers[:len(realmUsers)-1]
			isHasClient = true
		}
	}
	if !isHasClient {
		return fmt.Errorf("realm \"%s\" doesn't have user \"%s\" in Redis: %w", realmName, userName, errors_repo.ErrNotFound)
	}
	if err := mn.createRealmUsers(realmName, realmUsers, true); err != nil {
		return err
	}
	return nil
}

func (mn *RedisDataManager) UpdateUser(userName string, userValue []byte) (string, error) {
	oldUser, err := mn.GetUser(userName)
	if err != nil {
		return "", err
	}
	oldUserName := (*oldUser).GetUsername()
	oldUserId := (*oldUser).GetId()
	var newUser any
	if err := json.Unmarshal(userValue, &newUser); err != nil {
		return "", err
	}
	user := data.CreateUser(newUser)
	newUserName := user.GetUsername()
	newUserId := user.GetId()
	if newUserId != oldUserId || newUserName != oldUserName {
		// TODO каскадно обновлять информацию во всех realm где был этот user. И удалить сам user
	}
	userKey := sf.Format(userKeyTemplate, mn.namespace, newUserName)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, User, userKey, string(userValue)); err != nil {
		return "", err
	}
	return newUserName, nil
}
