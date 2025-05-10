package oauth

type UseCase interface {
	Provider() string
	ExchangeToken(authorizationCode, codeVerifier, redirectURI string) error
}
