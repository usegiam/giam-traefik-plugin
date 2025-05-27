package grafana

type MockRepo struct {
	User  *User
	Teams []*Team
	Err   error
}

func (g *MockRepo) GetUser(session string) (*User, error) {
	return g.User, g.Err
}

func (g *MockRepo) GetUserTeams(session string, userID int) ([]*Team, error) {
	return g.Teams, g.Err
}
