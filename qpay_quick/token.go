package qpay_quick

import "github.com/techpartners-asia/qpay-go/utils"

// Token is a qPay Quick access token; see [utils.Token] for the ownership
// contract. It is aliased rather than redefined so a token obtained from one
// qPay package can be handed to another without conversion.
type Token = utils.Token

// Amount is a qPay money field; see [utils.Amount] for why it is not a plain
// string. Aliased rather than redefined so an amount from one qPay package can
// be handed to another.
type Amount = utils.Amount

var (
	// ErrNoToken is returned when a call is made before [QPayQuick.SetToken].
	ErrNoToken = utils.ErrNoToken

	// ErrUnauthorized is returned when qPay rejects the installed token.
	ErrUnauthorized = utils.ErrUnauthorized
)

// tokenFrom converts a login or refresh response into a Token.
func tokenFrom(res qpayLoginResponse) Token {
	return Token{
		AccessToken:      res.AccessToken,
		RefreshToken:     res.RefreshToken,
		TokenType:        res.TokenType,
		ExpiresAt:        utils.TokenExpiresAt(res.ExpiresIn),
		RefreshExpiresAt: utils.TokenExpiresAt(res.RefreshExpiresIn),
		Scope:            res.Scope,
		SessionState:     res.SessionState,
	}
}
