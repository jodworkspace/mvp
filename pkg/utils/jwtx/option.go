package jwtx

type Option func(claims *Claims)

func WithIssuer(issuer string) Option {
	return func(claims *Claims) {
		claims.Issuer = issuer
	}
}

func WithSubject(subject string) Option {
	return func(claims *Claims) {
		claims.Subject = subject
	}
}

func WithAudience(audience string) Option {
	return func(claims *Claims) {
		claims.Audience = []string{audience}
	}
}
