package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/usegiam/giam-traefik-plugin/internal/errors"
	"github.com/usegiam/giam-traefik-plugin/pkg/log"
)

type repo struct {
	grafanaUrl string
	logger     *log.Logger
}

func NewRepo(grafanaUrl string, logger *log.Logger) Repo {
	return &repo{grafanaUrl: grafanaUrl, logger: logger}
}

func (r *repo) GetUser(session string) (*User, error) {
	req, err := http.NewRequest(http.MethodGet, r.grafanaUrl+"/api/user", nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "grafana_session",
		Value: session,
	})

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to send user request: %w", err)
	}
	defer resp.Body.Close()

	r.logger.Debugf("grafana get user resp: %v", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrFailedtoCommunicateWithGrafana
	}

	var user User

	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the user with payload: %w", err)
	}

	r.logger.Debugf("fetched user: %v", user)

	return &user, nil
}

func (r *repo) GetUserTeams(session string, userID int) ([]*Team, error) {
	req, err := http.NewRequest("GET", r.grafanaUrl+"/api/teams/search?userId="+strconv.Itoa(userID), nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "grafana_session",
		Value: session,
	})

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send teams request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrUserDoesntHaveAnyTeamAssigned
	}

	var response struct {
		Teams []*Team `json:"teams"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the teams with payload: %w", err)
	}

	r.logger.Debugf("fetched user teams")

	return response.Teams, nil
}
