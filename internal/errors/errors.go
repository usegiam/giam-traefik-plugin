package errors

import "errors"

var (
	ErrUnauthorized                   = errors.New("user is not authorized")
	ErrUnsupportedDatasource          = errors.New("unsupported datasource")
	ErrUserDoesntHaveAnyTeamAssigned  = errors.New("failed to fetch user teams, user might have not a team")
	ErrFailedtoCommunicateWithGrafana = errors.New("failed to communicate with grafana")
)
