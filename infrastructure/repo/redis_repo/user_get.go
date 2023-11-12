package redis_repo

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/wissance/Ferrum/data"
	sf "github.com/wissance/stringFormatter"
)

// GetUsersFromRealm function for getting all realm users
/* This function select all realm users (used by GetUserById) by constructing redis key from namespace and realm name
 * Probably in future this function could consume a lot of memory (if we would have a lot of users in a realm) probably we should limit amount of Users to fetch
 * This function works in two steps:
 *     1. Get all data.ExtendedIdentifier pairs id-name
 *     2. Get all User objects at once by key slices (every redis key for user combines from namespace, realm, username)
 * Parameters:
 *    - realmName - name of the realm
 * Returns slice of Users
 */
func (mn *RedisDataManager) GetUsersFromRealm(realmName string) []data.User {
	// TODO(UMV): possibly we should not use this method ??? what if we have 1M+ users .... ? think maybe it should be somehow optimized ...
	realmUsers := mn.getRealmUsers(realmName)
	if len(realmUsers) == 0 {
		return nil
	}

	// todo(UMV): probably we should organize batching here if we have many users i.e. 100K+
	userRedisKeys := make([]string, len(realmUsers))
	for i, ru := range realmUsers {
		userRedisKeys[i] = sf.Format(userKeyTemplate, mn.namespace, ru.Name)
	}

	// userFullDataRealmsKey := sf.Format(realmUsersFullDataKeyTemplate, mn.namespace, realmName)
	// this is wrong, we can't get rawUsers such way ...
	realmUsersData := getMultipleObjectFromRedis[interface{}](mn.redisClient, mn.ctx, mn.logger, RealmUsers, userRedisKeys)
	// getObjectsListFromRedis[interface{}](mn.redisClient, mn.ctx, mn.logger, RealmUsers, userFullDataRealmsKey)
	if len(realmUsers) != len(realmUsersData) {
		mn.logger.Error(sf.Format("Realm: \"{0}\" has users, that Redis does not have part of it", realmName))
	}

	if len(realmUsersData) == 0 {
		mn.logger.Error(sf.Format("Redis does not have all users that belong to Realm: \"{0}\"", realmName))
		return nil
	}
	userData := make([]data.User, len(realmUsersData))
	for i, u := range realmUsersData {
		userData[i] = data.CreateUser(u)
	}
	return userData
}

// GetUser function for getting realm user by username
/* This function constructs Redis key by pattern combines namespace, realm name and username (realmUsersKeyTemplate)
 * Parameters:
 *    - realm - pointer to realm
 *    - userName - name of user
 * Returns: User or nil
 */
func (mn *RedisDataManager) GetUser(userName string) (*data.User, error) {
	userKey := sf.Format(userKeyTemplate, mn.namespace, userName)
	rawUser := getObjectFromRedis[interface{}](mn.redisClient, mn.ctx, mn.logger, User, userKey)
	if rawUser == nil {
		return nil, fmt.Errorf("not found")
	}
	user := data.CreateUser(*rawUser)
	return &user, nil
}

// GetUserById function for getting realm user by userId
/* This function is more complex than GetUser, because we are using combination of realm name and username to store user data,
 * therefore this function extracts all realm users data and find appropriate by relation id-name after that it behaves like GetUser function
 * Parameters:
 *    - realm - pointer to realm
 *    - userId - identifier of searching user
 * Returns: User or nil
 */
func (mn *RedisDataManager) GetUserById(realmName string, userId uuid.UUID) (*data.User, error) {
	// TODO(sia) Переписать на более оптимальное
	// userKey := sf.Format(userKeyTemplate, mn.namespace, userId)
	var rawUser data.User
	userFound := false
	users := mn.GetUsersFromRealm(realmName)
	for _, u := range users {
		checkingUserId := u.GetId()
		if checkingUserId == userId {
			rawUser = u
			userFound = true
			break
		}
	}
	if !userFound {
		return nil, fmt.Errorf("not found")
	}

	return &rawUser, nil
}

func (mn *RedisDataManager) getRealmUser(realmName string, userName string) (*data.ExtendedIdentifier, error) {
	realmUsers := mn.getRealmUsers(realmName)
	if len(realmUsers) == 0 {
		return nil, fmt.Errorf("Not found")
	}

	var user data.ExtendedIdentifier
	userFound := false
	for _, rc := range realmUsers {
		if rc.Name == userName {
			userFound = true
			user = rc
			break
		}
	}

	if !userFound {
		return nil, fmt.Errorf("not found")
	}

	return &user, nil
}

func (mn *RedisDataManager) getRealmUsers(realmName string) []data.ExtendedIdentifier {
	userRealmsKey := sf.Format(realmUsersKeyTemplate, mn.namespace, realmName)
	realmUsers := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmUsers, userRealmsKey)
	if len(realmUsers) == 0 {
		mn.logger.Error(sf.Format("There are no users in realm: \"{0}\" in Redis", realmName))
		return nil
	}
	return realmUsers
}
