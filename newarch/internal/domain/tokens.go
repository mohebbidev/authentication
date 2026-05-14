package domain

type AccessTokens struct {
	Token
	UserID UserID
}

type VerificationToken struct {
	Token
	UserID UserID
}

type PasswordResetToken struct {
	Token
	UserID UserID
}