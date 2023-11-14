package in_memory

// TODO (sia) переименовать файл в inMemory

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"

	"github.com/wissance/Ferrum/internal/domain"
	"github.com/wissance/Ferrum/internal/utils/logging"
	"github.com/wissance/stringFormatter"
)

// FileDataManager is the simplest Data Storage without any dependencies, it uses single JSON file (it is users and clients RO auth server)
// This context type is extremely useful for simple systems
type InMemoryRepo struct {
	dataFile   string
	serverData ServerData
	logger     *logging.AppLogger
}

// PrepareFileDataContextUsingData initializes instance of FileDataManager and sets loaded data to serverData
/* This factory function creates initialize with data instance of  FileDataManager, error reserved for usage but always nil here
 * Parameters:
 *    serverData already loaded data.ServerData from Json file in memory
 * Returns: context and error (currently is nil)
 */
// func PrepareFileDataContextUsingData(serverData *ServerData) (*InMemoryRepo, error) {
// 	// todo(UMV): todo provide an error handling
// 	mn := &InMemoryRepo{serverData: *serverData}
// 	return mn, nil
// }

func NewInMemoryRepo(dataFile string, logger *logging.AppLogger) *InMemoryRepo {
	return &InMemoryRepo{
		dataFile:   dataFile,
		serverData: ServerData{},
		logger:     logger,
	}
}

func (mn *InMemoryRepo) UsingData(serverData ServerData) error {
	mn.serverData = serverData
	return nil
}

// loadData this function loads data from JSON file (dataFile) to serverData
func (mn *InMemoryRepo) LoadDataFromFile() error {
	rawData, err := os.ReadFile(mn.dataFile)
	if err != nil {
		mn.logger.Error(stringFormatter.Format("An error occurred during config file reading: {0}", err.Error()))
		return err
	}
	mn.serverData = ServerData{}
	if err = json.Unmarshal(rawData, &mn.serverData); err != nil {
		mn.logger.Error(stringFormatter.Format("An error occurred during data file unmarshal: {0}", err.Error()))
		return err
	}

	return nil
}

// GetRealm function for getting Realm by name
/* Searches for a realm with name realmName in serverData adn return it. Realm contains all related entities (clients, Users)
 * Parameters:
 *     - realmName - name of a realm
 * Returns: Realm or nil (if Realm isn't found0
 */
func (mn *InMemoryRepo) GetRealm(realmName string) (*domain.Realm, error) {
	for _, e := range mn.serverData.Realms {
		// case-sensitive comparison, myapp and MyApP are different realms
		if e.Name == realmName {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("not found") // TODO(ias)
}

// GetClient function for getting Realm Client by name
/* Searches for a client with name realmName in a realm. This function must be used after Realm was found.
 * Parameters:
 *     - realm - realm containing clients to search
 *     - name - name of a client
 * Returns: Client or nil (if Client isn't found0
 */
func (mn *InMemoryRepo) GetClientFromRealm(realmName string, clientName string) (*domain.Client, error) {
	realm, err := mn.GetRealm(realmName) // todo(sia): check
	if err != nil {
		return nil, err
	}

	for _, c := range realm.Clients {
		if c.Name == clientName {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("not found") // TODO(sia)
}

// GetUser function for getting Realm User by userName
/* Searches for a user with specified name in a realm.  This function must be used after Realm was found.
 * Parameters:
 *     - realm - realm containing users to search
 *     - userName - name of a user
 * Returns: realm user or nil
 */
func (mn *InMemoryRepo) GetUserFromRealm(realmName string, userName string) (*domain.User, error) {
	realm, err := mn.GetRealm(realmName) // todo(sia): check
	if err != nil {
		return nil, err
	}

	for _, u := range realm.Users {
		user := domain.NewUser(u)
		if user.GetUsername() == userName {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("not found") // TODO(sia)
}

// GetUserById function for getting Realm User by Id
/* same functions as GetUser but uses userId to search instead of username
 */
func (mn *InMemoryRepo) GetUserFromRealmById(realmName string, userId uuid.UUID) (*domain.User, error) {
	realm, err := mn.GetRealm(realmName) // todo(ias): check
	if err != nil {
		return nil, err
	}

	for _, u := range realm.Users {
		user := domain.NewUser(u)
		if user.GetId() == userId {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("not found") // TODO(sia)
}

// GetUsersForRealm function for getting all Realm User
/* This function get realm by name ant extract all its users
 * Parameters:
 *     - realmName - name of a realm
 * Returns: slice of users
 */
func (mn *InMemoryRepo) GetUsersFromRealm(realmName string) ([]domain.User, error) {
	realm, err := mn.GetRealm(realmName)
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, len(realm.Users))
	for i, u := range realm.Users {
		users[i] = domain.NewUser(u)
	}
	return users, nil
}
