package state

import "osto-auth-cli/internal/models"

type AppState struct {
	SessionToken string
	CurrentUser  *models.User
}

func (s *AppState) IsAuthenticated() bool {
	return s.SessionToken != ""
}

func (s *AppState) Clear() {
	s.SessionToken = ""
	s.CurrentUser = nil
}
