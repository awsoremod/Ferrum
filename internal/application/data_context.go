package application

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/wissance/Ferrum/internal/config"
	"github.com/wissance/Ferrum/internal/services"
	"github.com/wissance/Ferrum/internal/transport/rest"
	"github.com/wissance/Ferrum/internal/utils/logging"

	"github.com/wissance/Ferrum/internal/infrastructure/repo/in_memory"
	"github.com/wissance/Ferrum/internal/infrastructure/repo/redis"

	"github.com/wissance/stringFormatter"
)

// DataContext is a common interface to implement operations with authorization server entities (data.Realm, data.Client, data.User)
// now contains only set of Get methods, during implementation admin CLI should be expanded to create && update entities
type DataContext interface {
	// GetRealm(realmName string) (*domain.Realm, error)
	// GetUsersFromRealm(realmName string) ([]domain.User, error)

	services.RepoForTokenBasedSecurityService
	rest.RepoForWebApiContext
}

// PrepareContext is a factory function that creates instance of DataContext
/* This function creates instance of appropriate DataContext according to input arguments values, if dataSourceConfig is config.FILE function
 * creates instance of FileDataManager. For this type of context if dataFile is not nil and exists this function also provides data initialization:
 * loads all data (realms, clients and users) in a memory. If dataSourceCfg is config.REDIS this function creates instance of RedisDataManager
 * by calling CreateRedisDataManager function
 * Parameters:
 *     - dataSourceCfg configuration section related to DataSource
 *     - dataFile - data for initialization (this is using only when dataSourceCfg is config.FILE)
 *     - logger - logger instance
 * Return: new instance of DataContext and error (nil if there are no errors)
 */
func PrepareContext(dataSourceCfg *config.DataSourceConfig, dataFile *string, logger *logging.AppLogger) (DataContext, error) {
	var dc DataContext
	var err error
	switch dataSourceCfg.Type {
	case config.FILE:
		if dataFile == nil {
			err = errors.New("data file is nil")
			logger.Error(err.Error())
		}
		fmt.Println()
		absPath, pathErr := filepath.Abs(*dataFile)
		if pathErr != nil {
			// todo: umv: think what to do on error
			msg := stringFormatter.Format("An error occurred during attempt to get abs path of data file: {0}", err.Error())
			logger.Error(msg)
			err = pathErr
		}
		// init, load data in memory ...
		// mn := &in_memory.InMemoryRepo{dataFile: absPath, logger: logger}
		mn := in_memory.NewInMemoryRepo(absPath, logger)
		err = mn.LoadDataFromFile()
		if err != nil {
			// at least and think what to do further
			msg := stringFormatter.Format("An error occurred during data loading: {0}", err.Error())
			logger.Error(msg)
		}
		dc = mn

	case config.REDIS:
		if dataSourceCfg.Type == config.REDIS {
			dc, err = redis.NewRedisRepo(dataSourceCfg, logger)
		}
		// todo implement other data sources
	}

	return dc, err
}
