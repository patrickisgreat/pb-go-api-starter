package api

import (
	"slices"

	"github.com/twitchtv/twirp"

	"github.com/patrickisgreat/pb-go-api-starter/internal/generated/pb_go_api_starter/session"
	"github.com/patrickisgreat/pb-go-api-starter/internal/urn"
)

func isValidUrn(urnValue string, urnTypes []string) bool {
	if urnValue == "" {
		return false
	}

	u, err := urn.Parse(urnValue)
	if err != nil || (len(urnTypes) > 0 && !slices.Contains(urnTypes, u.Collection)) {
		return false
	}

	return true
}

func validateUserSession(userSession *session.UserSession) twirp.Error {
	if userSession.GetUserUrn().GetValue() != "" && !isValidUrn(userSession.GetUserUrn().GetValue(), []string{"users"}) {
		return twirp.InvalidArgumentError("userSession.userUrn", "invalid user urn")
	}

	return nil
}
