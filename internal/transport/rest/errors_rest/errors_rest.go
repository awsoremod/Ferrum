package errors_rest

const (
	RealmNotProviderMsg          = "You does not provided any realm"
	RealmDoesNotExistsTemplate   = "Realm \"{0}\" does not exists"
	BadBodyForTokenGenerationMsg = "Bad body for token generation, see documentations"
	InvalidClientMsg             = "Invalid client"
	InvalidClientCredentialDesc  = "Invalid client credentials"
	InvalidUserCredentialsMsg    = "invalid grant"
	InvalidUserCredentialsDesc   = "Invalid user credentials"
	InvalidRequestMsg            = "Invalid request"
	InvalidRequestDesc           = "Token not provided"
	InvalidTokenMsg              = "Invalid token"
	InvalidTokenDesc             = "Token verification failed"
	TokenIsNotActive             = "Token is not active"
)
