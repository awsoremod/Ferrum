package rest

import (
	"github.com/google/uuid"

	"github.com/wissance/Ferrum/internal/domain"
	"github.com/wissance/Ferrum/internal/services"
	"github.com/wissance/Ferrum/internal/transport/dto"
	"github.com/wissance/Ferrum/internal/utils/logging"
)

type RepoForWebApiContext interface {
	GetRealm(realmName string) (*domain.Realm, error)
	GetUserFromRealmById(realmName string, userId uuid.UUID) (*domain.User, error)
}

// SecurityService is an interface that implements all checks and manipulation with sessions data
type ServiceSecurityForWebApiContext interface {
	// Validate checks whether provided tokenIssueData could be used for token generation or not
	Validate(tokenIssueData *dto.TokenGenerationData, realm *domain.Realm) *domain.OperationError
	// CheckCredentials validates provided in tokenIssueData pairs of clientId+clientSecret and username+password
	CheckCredentials(tokenIssueData *dto.TokenGenerationData, realm *domain.Realm) *domain.OperationError
	// GetCurrentUserByName return CurrentUser data by name
	GetCurrentUserByName(realm *domain.Realm, userName string) *domain.User
	// GetCurrentUserById return CurrentUser data by шв
	GetCurrentUserById(realm *domain.Realm, userId uuid.UUID) *domain.User
	// StartOrUpdateSession starting new session on new successful token issue request or updates existing one with new request with valid token
	StartOrUpdateSession(realm string, userId uuid.UUID, duration int, refresh int) uuid.UUID
	// AssignTokens this function creates relation between userId and issued tokens (access and refresh)
	AssignTokens(realm string, userId uuid.UUID, accessToken *string, refreshToken *string)
	// GetSession returns user session data
	GetSession(realm string, userId uuid.UUID) *domain.UserSession
	// GetSessionByAccessToken returns session data by access token
	GetSessionByAccessToken(realm string, token *string) *domain.UserSession
	// GetSessionByRefreshToken returns session data by access token
	GetSessionByRefreshToken(realm string, token *string) *domain.UserSession
	// CheckSessionAndRefreshExpired checks is user tokens expired or not (could user use them or should get new ones)
	CheckSessionAndRefreshExpired(realm string, userId uuid.UUID) (bool, bool)
}

// WebApiContext is a central Application logic processor manages from Web via HTTP/HTTPS
type WebApiContext struct {
	Address        string
	Schema         string
	DataProvider   RepoForWebApiContext
	Security       ServiceSecurityForWebApiContext
	TokenGenerator *services.JwtGeneratorService
	Logger         *logging.AppLogger
}
