package main

import (
	"fmt"
	"log"

	"github.com/wissance/Ferrum/api/admin/cli/config_cli"
	"github.com/wissance/Ferrum/api/admin/cli/domain"
	"github.com/wissance/Ferrum/data"
	"github.com/wissance/Ferrum/logging"
	"github.com/wissance/Ferrum/managers"

	sf "github.com/wissance/stringFormatter"
)

type IManagerForCli interface {
	GetRealm(realmName string) (*data.Realm, error)
	GetClient(clientName string) (*data.Client, error)
	GetRealmClient(realmName string, clientName string) (*data.ExtendedIdentifier, error)
	GetUser(userName string) (*data.User, error)
	GetRealmUser(realmName string, clientName string) (*data.ExtendedIdentifier, error)

	CreateRealm(realmValue []byte) (*data.Realm, error)
	CreateClient(clientValue []byte) (*data.Client, error)
	AddClientToRealm(realmName string, clientName string) error
	CreateUser(userValue []byte) (string, error)
	AddUserToRealm(realmName string, userName string) error

	DeleteRealm(realmName string) error
	DeleteClient(clientName string) error
	DeleteRealmClient(realmName string, clientName string) error
	DeleteUser(userName string) error
	DeleteRealmUser(realmName string, userName string) error

	UpdateClient(clientName string, clientValue []byte) (*data.Client, error)
	UpdateUser(userName string, userValue []byte) (string, error)
	UpdateRealm(realmName string, realmValue []byte) (*data.Realm, error)
}

func main() {
	cfg, err := config_cli.NewConfig()
	if err != nil {
		log.Fatalf("NewConfig failed: %s", err)
	}

	var manager IManagerForCli
	{
		logger := logging.CreateLogger(&cfg.LoggingConfig)
		redisManager, err := managers.NewRedisDataManager(&cfg.DataSourceConfig, logger)
		if err != nil {
			log.Fatalf("CreateRedisDataManager failed: %s", err)
		}
		manager = redisManager
	}

	switch cfg.Operation {
	case domain.GetOperation:
		if cfg.Resource_id == "" {
			log.Fatal(sf.Format("Bad Resource_id: \"{0}\"", cfg.Resource_id))
		}
		switch cfg.Resource {
		case domain.ClientResource:
			if cfg.Params == "" {
				client, err := manager.GetClient(cfg.Resource_id)
				if err != nil {
					log.Fatal() // TODO redis not have
				}
				fmt.Println(*client)
			} else {
				clientIdAndName, err := manager.GetRealmClient(cfg.Params, cfg.Resource_id)
				if err != nil {
					log.Fatal(err) // TODO Not have in realm
				}
				fmt.Println(*clientIdAndName)
			}

		case domain.UserResource:
			if cfg.Params == "" {
				user, err := manager.GetUser(cfg.Resource_id)
				if err != nil {
					log.Fatal() // TODO redis not have
				}
				fmt.Println(*user)
			} else {
				userIdAndName, err := manager.GetRealmUser(cfg.Params, cfg.Resource_id)
				if err != nil {
					log.Fatal(err) // TODO Not have in realm
				}
				fmt.Println(*userIdAndName)
			}

		case domain.RealmResource:
			realm, err := manager.GetRealm(cfg.Resource_id)
			if err != nil {
				log.Fatal(sf.Format("Realm: \"{0}\" doesn't exist", cfg.Resource_id)) // TODO
			}
			fmt.Println(*realm)
		}

		return
	case domain.CreateOperation:
		switch cfg.Resource {
		case domain.ClientResource:
			if cfg.Params == "" {
				if len(cfg.Value) == 0 {
					log.Fatal(sf.Format("Bad Value: len zero"))
				}
				client, err := manager.CreateClient(cfg.Value)
				if err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("Client: \"{0}\" successfully created", client.Name))

			} else {
				if cfg.Resource_id == "" {
					log.Fatal(sf.Format("Bad Resource_id: \"{0}\"", cfg.Resource_id))
				}
				if err := manager.AddClientToRealm(cfg.Params, cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("Client: \"{0}\" successfully added to Realm: \"{1}\"", cfg.Resource_id, cfg.Params))
			}

		case domain.UserResource:
			if cfg.Params == "" {
				if len(cfg.Value) == 0 {
					log.Fatal(sf.Format("Bad Value: len zero"))
				}
				userName, err := manager.CreateUser(cfg.Value)
				if err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("User: \"{0}\" successfully created", userName))

			} else {
				if cfg.Resource_id == "" {
					log.Fatal(sf.Format("Bad Resource_id: \"{0}\"", cfg.Resource_id))
				}
				if err := manager.AddUserToRealm(cfg.Params, cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("User: \"{0}\" successfully added to Realm: \"{1}\"", cfg.Resource_id, cfg.Params))
			}

		case domain.RealmResource:
			if len(cfg.Value) == 0 {
				log.Fatal(sf.Format("Bad Value: len zero"))
			}
			// создает клиентов и пользователей, создает новые realmClients и realmUsers, создает realm
			realm, err := manager.CreateRealm(cfg.Value)
			if err != nil {
				log.Fatal("%s", err)
			}
			fmt.Println(sf.Format("Realm: \"{0}\" successfully created", realm.Name))
			return
		}

		return
	case domain.DeleteOperation:
		if cfg.Resource_id == "" {
			log.Fatal(sf.Format("Bad Resource_id: \"{0}\"", cfg.Resource_id))
		}
		switch cfg.Resource {
		case domain.ClientResource:
			if cfg.Params == "" {
				if err := manager.DeleteClient(cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("Client: \"{0}\" successfully deleted", cfg.Resource_id))
			} else {
				// Удаляет клиента из realmClients. Удаление самого клиента не происходит
				if err := manager.DeleteRealmClient(cfg.Params, cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("Client: \"{0}\" successfully deleted in Realm: \"{1}\"", cfg.Resource_id, cfg.Params))
			}

		case domain.UserResource:
			if cfg.Params == "" {
				if err := manager.DeleteUser(cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("User: \"{0}\" successfully deleted", cfg.Resource_id))
			} else {
				// Удаляет user из realmUsers. Удаление самого клиента не происходит
				if err := manager.DeleteRealmUser(cfg.Params, cfg.Resource_id); err != nil {
					log.Fatalf("%s", err)
				}
				fmt.Println(sf.Format("User: \"{0}\" successfully deleted in Realm: \"{1}\"", cfg.Resource_id, cfg.Params))
			}

		case domain.RealmResource:
			// Удаляет realmClients и realmUsers и realm. Удаление самих client и user не происходит.
			if err := manager.DeleteRealm(cfg.Resource_id); err != nil {
				log.Fatalf("%s", err)
			}
			fmt.Println(sf.Format("Realm: \"{0}\" successfully deleted", cfg.Resource_id))
		}

		return
	case domain.UpdateOperation:
		switch cfg.Resource {
		case domain.ClientResource:
			client, err := manager.UpdateClient(cfg.Resource_id, cfg.Value)
			if err != nil {
				log.Fatalf("%s", err)
			}
			fmt.Println(sf.Format("Client: \"{0}\" successfully updated", client.Name))

		case domain.UserResource:
			userName, err := manager.UpdateUser(cfg.Resource_id, cfg.Value)
			if err != nil {
				log.Fatalf("%s", err)
			}
			fmt.Println(sf.Format("User: \"{0}\" successfully updated", userName, cfg.Params))

		case domain.RealmResource:

			// if err := redisManager.DeleteRealm(cfg.Resource_id); err != nil {
			// 	log.Fatalf("%s", err)
			// }
			fmt.Println(sf.Format("Realm: \"{0}\" successfully updated", cfg.Resource_id))
		}

		return
	case "change_password":
		fmt.Println("change_password")

	case "reset_password ":
		fmt.Println("reset_password")

	default:
		log.Fatalf("Bad Operation") // TODO
	}
}
