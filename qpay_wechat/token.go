package qpay_wechat

import "github.com/techpartners-asia/qpay-go/utils"

// Token is a qPay access token; see [utils.Token] for the ownership contract.
// It is aliased rather than redefined so a token obtained from one qPay
// package can be handed to another without conversion.
type Token = utils.Token

var (
	// ErrNoToken is returned when a call is made before [QPayAuth.SetToken].
	ErrNoToken = utils.ErrNoToken

	// ErrUnauthorized is returned when qPay rejects the installed token.
	ErrUnauthorized = utils.ErrUnauthorized
)

// tokenFrom converts a login or refresh response into a Token. This package
// types its expiry fields as int rather than int64; the semantics are the same.
func tokenFrom(res qpayLoginResponse) Token {
	return Token{
		AccessToken:      res.AccessToken,
		RefreshToken:     res.RefreshToken,
		TokenType:        res.TokenType,
		ExpiresAt:        utils.TokenExpiresAt(int64(res.ExpiresIn)),
		RefreshExpiresAt: utils.TokenExpiresAt(int64(res.RefreshExpiresIn)),
		Scope:            res.Scope,
		SessionState:     res.SessionState,
	}
}
