package main

import (
	"fmt"
	"log"

	cli_config "github.com/wissance/Ferrum/api/admin/cli/config"
	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/logging"
	"github.com/wissance/Ferrum/managers"

	sf "github.com/wissance/stringFormatter"
)

func main() {
	cfg, _ := cli_config.NewConfig()
	// TODO мб перенести в valide в config
	if cfg.Resource != "realm" && cfg.Resource != "client" && cfg.Resource != "user" {
		log.Fatalf(sf.Format("bad Resource: \"{0}\"", cfg.Resource)) // TODO
	}
	if cfg.Namespace == "" {
		log.Fatalf(sf.Format("bad Namespace: \"{0}\"", cfg.Namespace)) // TODO
	}
	if cfg.Resource_id == "" {
		log.Fatalf(sf.Format("bad Resource_id: \"{0}\"", cfg.Resource_id)) // TODO
	}

	logger := logging.CreateLogger(&cfg.LoggingConfig)
	redisManager, err := managers.CreateRedisDataManager(&cfg.DataSourceConfig, logger) // TODO вынести в интерфейс DataContext
	if err != nil {
		log.Fatal(err)
	}

	// cfg.Params уточнить как указывается
	switch cfg.Operation {
	case "get":
		var realm *data.Realm
		if cfg.Resource == "realm" {
			realm = redisManager.GetRealm(cfg.Resource_id)
		} else {
			if cfg.Params == "" {
				fmt.Println("You need to specify Realm name in --params.")
				return
			}
			realm = redisManager.GetRealm(cfg.Params) // TODO передалть возврат с clients и users
		}
		if realm == nil {
			fmt.Println(sf.Format("Realm: \"{0}\" doesn't exist", cfg.Resource_id))
			return
			// wCtx.Logger.Debug("realm doesn't exist")
			// result = dto.ErrorDetails{Msg: stringFormatter.Format(errors.RealmDoesNotExistsTemplate, realm)}
		}

		switch cfg.Resource {
		case "realm":
			fmt.Println(*realm)
		case "client":
			client := redisManager.GetClient(realm, cfg.Resource_id)
			if client == nil {
				fmt.Println(sf.Format("Client: \"{0}\" in Realm: \"{1}\" doesn't exist", cfg.Resource_id, realm.Name))
				return
			}
			fmt.Println(*client)
		case "user":
			user := redisManager.GetUser(realm, cfg.Resource_id)
			if user == nil {
				fmt.Println(sf.Format("User: \"{0}\" in Realm: \"{1}\" doesn't exist", cfg.Resource_id, realm.Name))
				return
			}
			fmt.Println(*user)
		}
		return

	case "create":
		switch cfg.Resource {
		case "realm":
			// TODO вопрос, разрешать ли перезапись
			if err := redisManager.CreateRealm(cfg.Resource_id, cfg.Value); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Realm successfully created.") // TODO
		case "client":

		case "user":

		}
		return

	case "update":
		fmt.Println("client")

	case "delete":
		fmt.Println("user")

	case "change_password":
		fmt.Println("change_password")

	case "reset_password ":
		fmt.Println("reset_password")

	default:
		log.Fatalf("bad Operation") // TODO
	}
}
