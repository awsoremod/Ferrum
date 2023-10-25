package managers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wissance/Ferrum/config"
	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/logging"
	sf "github.com/wissance/stringFormatter"
)

// This set of const of a templates to all data storing in Redis it contains prefix - a namespace {0}
const (
	userKeyTemplate         = "{0}.user_{1}"
	realmKeyTemplate        = "{0}.realm_{1}"
	realmClientsKeyTemplate = "{0}.realm_{1}_clients"
	clientKeyTemplate       = "{0}.client_{1}"
	realmUsersKeyTemplate   = "{0}.realm_{1}_users"
	// realmUsersFullDataKeyTemplate = "{0}.realm_{1}_users_full_data"
)

type objectType string

const (
	Realm        objectType = "realm"
	RealmClients            = "realm clients"
	RealmUsers              = "realm users"
	Client                  = "client"
	User                    = "user"
)

const defaultNamespace = "fe"

// RedisDataManager is a redis client
/*
 * Redis Data Manager is a service class for managing authorization server data in Redis
 * There are following store Rules:
 * 1. Realms (data.Realm) in Redis storing separately from Clients && Users, every Realm stores in Redis by key forming from template && Realm name
 *    i.e. if we have Realm with name "wissance" it could be accessed by key fe_realm_wissance (realmKeyTemplate)
 * 2. Realm Clients ([]data.ExtendedIdentifier) storing in Redis by key forming from template, Realm with name wissance has array of clients id by key
 *    fe_realm_wissance_clients (realmClientsKeyTemplate)
 * 3. Every Client (data.Client) stores separately by key forming from client id (different realms could have clients with same name but in different realm,
 *    Client Name is unique only in Realm) and template clientKeyTemplate, therefore realm with pair (ID: 6e09faca-1004-11ee-be56-0242ac120002 Name: homeApp)
 *    could be received by key - fe_client_6e09faca-1004-11ee-be56-0242ac120002
 * 4. Every User in Redis storing by it own key forming by userId + template (userKeyTemplate) -> i.e. user with id 6dee45ee-1056-11ee-be56-0242ac120002 stored
 *    by key fe_user_6dee45ee-1056-11ee-be56-0242ac120002
 * 5. Client to Realm and User to Realm relation stored by separate keys forming using template and realm name, these relations stores array of data.ExtendedIdentifier
 *    that wires together Realm Name with User.ID and User.Name.
 *    IMPORTANT NOTES:
 *    1. When save Client or User don't forget to save relations in Redis too (see templates realmClientsKeyTemplate && realmUsersKeyTemplate)
 *    2. When add/modify new or existing user don't forget to update realmUsersFullDataKeyTemplate maybe this collection will be removed in future but currently
 *       we have it.
 */
type RedisDataManager struct {
	namespace   string
	redisOption *redis.Options
	redisClient *redis.Client
	logger      *logging.AppLogger
	ctx         context.Context
}

// CreateRedisDataManager is factory function for instance of RedisDataManager creation and return as interface DataContext
/* Simply creates instance of RedisDataManager and initializes redis client, this function requires config.Namespace to be set up in configs, otherwise
 * defaultNamespace is using
 * Parameters:
 *     - dataSourceCfg contains Redis specific settings in Options map (see allowed keys of map in config.DataSourceConnOption)
 *     - logger - initialized logger instance
 */
func CreateRedisDataManager(dataSourceCfg *config.DataSourceConfig, logger *logging.AppLogger) (DataContext, error) {
	// todo(UMV): todo provide an error handling
	opts := buildRedisConfig(dataSourceCfg, logger)
	rClient := redis.NewClient(opts)
	namespace, ok := dataSourceCfg.Options[config.Namespace]
	if !ok || len(namespace) == 0 {
		namespace = defaultNamespace
	}
	mn := &RedisDataManager{
		logger: logger, redisOption: opts, redisClient: rClient, ctx: context.Background(),
		namespace: namespace,
	}
	dc := DataContext(mn)
	return dc, nil
}

// GetRealm function for getting realm by name
/* This function constructs Redis key by pattern combines namespace and realm name (realmKeyTemplate). Unlike from FILE provider
 * Realm stored in Redis does not have Clients and Users inside Realm itself, these objects must be queried separately.
 * Parameters:
 *     - realmName name of a realm
 * Returns: realms or nil (if realm was not found)
 */
func (mn *RedisDataManager) GetRealm(realmName string) *data.Realm {
	realmKey := sf.Format(realmKeyTemplate, mn.namespace, realmName)
	realm := getObjectFromRedis[data.Realm](mn.redisClient, mn.ctx, mn.logger, Realm, realmKey)
	if realm == nil {
		return nil
	}

	// should get realms too
	// if realms were stored without clients (we expected so), get clients related to realm and assign here
	if len(realm.Clients) == 0 {
		realm.Clients = mn.GetRealmClients(realmName)
	}
	return realm
}

// GetClient function for get realm client by name
/* This function constructs Redis key by pattern combines namespace and realm name (realmClientsKeyTemplate)
 * Parameters:
 *     - realm - pointer to realm
 *     - name - client name
 * Returns: client or nil
 */
func (mn *RedisDataManager) GetClient(realm *data.Realm, name string) *data.Client {
	realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realm.Name)
	realmClients := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmClients, realmClientsKey)
	if len(realmClients) == 0 {
		mn.logger.Error(sf.Format("There are no clients for realm: \"{0}\" in Redis, BAD data config", realm.Clients))
		return nil
	}

	realmHasClient := false
	var clientId data.ExtendedIdentifier
	for _, rc := range realmClients {
		if rc.Name == name {
			realmHasClient = true
			clientId = rc
			break
		}
	}
	if !realmHasClient {
		// TODO (sia) debug не показывается
		mn.logger.Debug(sf.Format("Realm: \"{0}\" doesn't have client : \"{1}\" in Redis", realm.Name, name))
		return nil
	}

	clientKey := sf.Format(clientKeyTemplate, mn.namespace, clientId.Name)
	client := getObjectFromRedis[data.Client](mn.redisClient, mn.ctx, mn.logger, Client, clientKey)
	if client == nil {
		mn.logger.Error(sf.Format("Realm: \"{0}\" has client: \"{1}\", that Redis does not have", realm.Name, name))
		return nil
	}
	return client
}

// GetUser function for getting realm user by username
/* This function constructs Redis key by pattern combines namespace, realm name and username (realmUsersKeyTemplate)
 * Parameters:
 *    - realm - pointer to realm
 *    - userName - name of user
 * Returns: User or nil
 */
func (mn *RedisDataManager) GetUser(realm *data.Realm, userName string) *data.User {
	userRealmsKey := sf.Format(realmUsersKeyTemplate, mn.namespace, realm.Name)
	realmUsers := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmUsers, userRealmsKey)
	if len(realmUsers) == 0 {
		mn.logger.Error(sf.Format("There are no users in realm: \"{0}\" in Redis", realm.Name))
		return nil
	}

	var extendedUserId data.ExtendedIdentifier
	userFound := false
	for _, rc := range realmUsers {
		if rc.Name == userName {
			userFound = true
			extendedUserId = rc
			break
		}
	}

	if !userFound {
		mn.logger.Debug(sf.Format("User with name: \"{0}\" was not found for realm: \"{1}\"", userName, realm.Name))
		return nil
	}

	userKey := sf.Format(userKeyTemplate, mn.namespace, extendedUserId.Name)
	rawUser := getObjectFromRedis[interface{}](mn.redisClient, mn.ctx, mn.logger, User, userKey)
	if rawUser == nil {
		mn.logger.Error(sf.Format("Realm: \"{0}\" has user: \"{1}\", that Redis does not have", realm.Name, userName))
		return nil
	}
	user := data.CreateUser(*rawUser)
	return &user
}

// GetUserById function for getting realm user by userId
/* This function is more complex than GetUser, because we are using combination of realm name and username to store user data,
 * therefore this function extracts all realm users data and find appropriate by relation id-name after that it behaves like GetUser function
 * Parameters:
 *    - realm - pointer to realm
 *    - userId - identifier of searching user
 * Returns: User or nil
 */
func (mn *RedisDataManager) GetUserById(realm *data.Realm, userId uuid.UUID) *data.User {
	// userKey := sf.Format(userKeyTemplate, mn.namespace, userId)
	var rawUser data.User
	userFound := false
	users := mn.GetRealmUsers(realm.Name)
	for _, u := range users {
		checkingUserId := u.GetId()
		if checkingUserId == userId {
			rawUser = u
			userFound = true
			break
		}
	}
	if !userFound {
		return nil
	}

	return &rawUser
}

// GetRealmUsers function for getting all realm users
/* This function select all realm users (used by GetUserById) by constructing redis key from namespace and realm name
 * Probably in future this function could consume a lot of memory (if we would have a lot of users in a realm) probably we should limit amount of Users to fetch
 * This function works in two steps:
 *     1. Get all data.ExtendedIdentifier pairs id-name
 *     2. Get all User objects at once by key slices (every redis key for user combines from namespace, realm, username)
 * Parameters:
 *    - realmName - name of the realm
 * Returns slice of Users
 */
func (mn *RedisDataManager) GetRealmUsers(realmName string) []data.User {
	// TODO(UMV): possibly we should not use this method ??? what if we have 1M+ users .... ? think maybe it should be somehow optimized ...
	userRealmsKey := sf.Format(realmUsersKeyTemplate, mn.namespace, realmName)

	realmUsers := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmUsers, userRealmsKey)
	if len(realmUsers) == 0 {
		mn.logger.Error(sf.Format("There are no users in realm: \"{0}\" in Redis", realmName))
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

// GetRealmClients function for getting all realm clients
/* This function gets all realm client.
 * This function works in two steps:
 *     1. Get all data.ExtendedIdentifier pairs id-name
 *     2. Get all Client objects at once by key slices (every redis key for client combines from namespace, realm and client name)
 * Parameters:
 *    - realmName - name of the realm
 * Returns slice of Clients or nil
 */
func (mn *RedisDataManager) GetRealmClients(realmName string) []data.Client {
	realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realmName)
	realmClients := getObjectsListFromRedis[data.ExtendedIdentifier](mn.redisClient, mn.ctx, mn.logger, RealmClients, realmClientsKey)
	if len(realmClients) == 0 {
		mn.logger.Error(sf.Format("There are no clients for realm: \"{0}\" in Redis, BAD data config", realmName))
		return nil
	}
	clients := make([]data.Client, len(realmClients))
	for i, rc := range realmClients {
		// todo(UMV) get all them at once
		clientKey := sf.Format(clientKeyTemplate, mn.namespace, rc.Name)
		client := getObjectFromRedis[data.Client](mn.redisClient, mn.ctx, mn.logger, Client, clientKey)
		if client == nil {
			mn.logger.Error(sf.Format("Realm: \"{0}\" has client: \"{1}\", that Redis does not have", realmName, rc.Name))
			return nil
		}
		clients[i] = *client
	}

	return clients
}

func (mn *RedisDataManager) CreateRealm(realmName string, realmValue []byte) (*data.Realm, error) {
	// TODO не забыть про аналог транзакции, что делать если ошибка
	var realm data.Realm
	err := json.Unmarshal(realmValue, &realm)
	if err != nil {
		mn.logger.Error(sf.Format("An error occurred during Realm: \"{0}\" unmarshall", realmName))
		return nil, err
	}

	if len(realm.Clients) != 0 {
		bytesClients, err := json.Marshal(realm.Clients)
		if err != nil {
			return nil, err
		}

		// TODO возможно нужно проверять, что есть какие-то поля у clients

		realmClientsKey := sf.Format(realmClientsKeyTemplate, mn.namespace, realmName)
		redisIntCmd := mn.redisClient.Del(mn.ctx, realmClientsKey) // TODO уточнить
		if redisIntCmd.Err() != nil {
			return nil, redisIntCmd.Err()
		}
		redisIntCmd = mn.redisClient.RPush(mn.ctx, realmClientsKey, string(bytesClients))
		if redisIntCmd.Err() != nil {
			return nil, redisIntCmd.Err()
		}

		for _, client := range realm.Clients {
			bytesClient, err := json.Marshal(client)
			if err != nil {
				return nil, err
			}
			clientKey := sf.Format(clientKeyTemplate, mn.namespace, realmName)
			if err := setString(mn.redisClient, mn.ctx, mn.logger, Client, clientKey, string(bytesClient)); err != nil {
				return nil, err
			}
		}

		realm.Clients = []data.Client{}
	}

	if len(realm.Users) != 0 {
		// TODO тоже самое сделать

		realm.Users = []any{}
	}

	realmKey := sf.Format(realmKeyTemplate, mn.namespace, realmName)
	if err := setString(mn.redisClient, mn.ctx, mn.logger, Realm, realmKey, string(realmValue)); err != nil {
		return nil, err
	}
	// TODO

	return &realm, nil // TODO нет смысла возвращать реалм без client и user. Нужно наверное делать глубокую копию
}

func (mn *RedisDataManager) CreateRealmClients() error {
}

func setString(redisClient *redis.Client, ctx context.Context, logger *logging.AppLogger,
	objName objectType, objKey string, objValue string,
) error {
	statusCmd := redisClient.Set(ctx, objKey, objValue, 0)
	if statusCmd.Err() != nil {
		logger.Warn(sf.Format("An error occurred during set {0}: \"{1}\": \"{2}\" from Redis server", objName, objKey, objValue))
		return statusCmd.Err()
	}
	return nil
}

// getObjectFromRedis is a method that DOESN'T work with List type object, only a String object type
func getObjectFromRedis[T any](redisClient *redis.Client, ctx context.Context, logger *logging.AppLogger,
	objName objectType, objKey string,
) *T {
	redisCmd := redisClient.Get(ctx, objKey)
	if redisCmd.Err() != nil {
		logger.Warn(sf.Format("An error occurred during fetching {0}: \"{1}\" from Redis server", objName, objKey))
		return nil
	}

	var obj T
	jsonBin := []byte(redisCmd.Val())
	err := json.Unmarshal(jsonBin, &obj)
	if err != nil {
		logger.Error(sf.Format("An error occurred during {0} : \"{1}\" unmarshall", objName, objKey))
		return nil
	}
	return &obj
}

// getObjectFromRedis is a method that DOESN'T work with List type object, only a String object type
func getMultipleObjectFromRedis[T any](redisClient *redis.Client, ctx context.Context, logger *logging.AppLogger,
	objName objectType, objKey []string,
) []T {
	redisCmd := redisClient.MGet(ctx, objKey...)
	if redisCmd.Err() != nil {
		// todo(UMV): print when this will be done https://github.com/Wissance/stringFormatter/issues/14
		logger.Warn(sf.Format("An error occurred during fetching {0}: from Redis server", objName))
		return nil
	}

	raw := redisCmd.Val()
	if len(raw) == 0 {
		return nil
	}
	result := make([]T, len(raw))
	var unMarshalledRaw interface{}
	for i, v := range raw {
		err := json.Unmarshal([]byte(v.(string)), &unMarshalledRaw)
		if err != nil {
			logger.Error(sf.Format("An error occurred during {0} : \"{1}\" unmarshall", objName, objKey))
			return nil
		}
		result[i] = unMarshalledRaw.(T)
	}
	return result
}

// this functions gets object that stored as a LIST Object type
func getObjectsListFromRedis[T any](redisClient *redis.Client, ctx context.Context, logger *logging.AppLogger,
	objName objectType, objKey string,
) []T {
	redisCmd := redisClient.LRange(ctx, objKey, 0, -1)
	if redisCmd.Err() != nil {
		logger.Warn(sf.Format("An error occurred during fetching {0}: \"{1}\" from Redis server", objName, objKey))
		return nil
	}

	// var obj T
	items := redisCmd.Val()
	if len(items) == 0 {
		return nil
	}
	var result []T
	var portion []T
	for _, rawVal := range items {
		jsonBin := []byte(rawVal)
		err := json.Unmarshal(jsonBin, &portion) // already contains all SLICE in one object
		if err != nil {
			logger.Error(sf.Format("An error occurred during {0} : \"{1}\" unmarshall", objName, objKey))
			return nil
		}
		result = append(result, portion...)
	}
	return result
}

// buildRedisConfig builds redis.Options from map of values by known in config package set of keys
func buildRedisConfig(dataSourceCfd *config.DataSourceConfig, logger *logging.AppLogger) *redis.Options {
	dbNum, err := strconv.Atoi(dataSourceCfd.Options[config.DbNumber])
	if err != nil {
		logger.Error(sf.Format("can't be because we already called Validate(), but in any case: parsing error: {0}", err.Error()))
		return nil
	}
	opts := redis.Options{
		Addr: dataSourceCfd.Source,
		DB:   dbNum,
	}
	// passing credentials if we have it
	if dataSourceCfd.Credentials != nil {
		opts.Username = dataSourceCfd.Credentials.Username
		opts.Password = dataSourceCfd.Credentials.Password
	}
	// passing TLS if we have it
	val, ok := dataSourceCfd.Options[config.UseTls]
	if ok {
		useTls, parseErr := strconv.ParseBool(val)
		if parseErr == nil && useTls {
			opts.TLSConfig = &tls.Config{}
			val, ok = dataSourceCfd.Options[config.InsecureTls]
			if ok {
				inSecTls, parseInSecValErr := strconv.ParseBool(val)
				if parseInSecValErr == nil {
					opts.TLSConfig.InsecureSkipVerify = inSecTls
				}
			}
		}
	}

	return &opts
}
