package main

import (
	"fmt"
	"log"

	cli_config "github.com/wissance/Ferrum/api/admin/cli/config"
	"github.com/wissance/Ferrum/logging"
	"github.com/wissance/Ferrum/managers"
)

func main() {
	cfg, _ := cli_config.NewConfig()
	// TODO мб перенести в valide в config
	if cfg.Resource != "realm" && cfg.Resource != "client" && cfg.Resource != "user" {
		log.Fatalf("bad Resource") // TODO
	}
	if cfg.Namespace == "" {
		log.Fatalf("bad Namespace") // TODO
	}
	if cfg.Resource_id == "" {
		log.Fatalf("bad Resource_id") // TODO
	}

	logger := logging.CreateLogger(&cfg.LoggingConfig)
	redisManager, err := managers.CreateRedisDataManager(&cfg.DataSourceConfig, logger) // TODO вынести в интерфейс DataContext
	if err != nil {
		log.Fatal(err)
	}

	// cfg.Params уточнить как указывается
	switch cfg.Operation {
	case "get":
		realm := redisManager.GetRealm(cfg.Params) // TODO передалть возврат с clients и users
		if realm == nil {
			fmt.Println("realm doesn't exist")
			return
			// wCtx.Logger.Debug("realm doesn't exist")
			// result = dto.ErrorDetails{Msg: stringFormatter.Format(errors.RealmDoesNotExistsTemplate, realm)}
		}
		switch cfg.Resource {
		case "realm":
			fmt.Println(realm)
		case "client":
			client := redisManager.GetClient(realm, cfg.Resource_id)
			if client == nil {
				fmt.Println("client doesn't exist")
				return
			}
			fmt.Println(client)
		case "user":
			user := redisManager.GetUser(realm, cfg.Resource_id)
			if user == nil {
				fmt.Println("user doesn't exist")
				return
			}
			fmt.Println(user)
		}
		return

	case "create":
		fmt.Println("realm")
	case "update":
		fmt.Println("client")
	case "delete":
		fmt.Println("user")
	default:
		log.Fatalf("bad Operation") // TODO
	}
}
