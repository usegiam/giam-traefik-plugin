package grafana

type Repo interface {
	GetUser(session string) (*User, error)
	GetUserTeams(session string, userID int) ([]*Team, error)
}

type QueryReq struct {
	Queries []interface{} `json:"queries"`
	From    string        `json:"from"`
	To      string        `json:"to"`
}

type SeriesReq struct {
	Series []map[string]string `json:"data"`
	Status string              `json:"status"`
}

type LabelValuesReq struct {
	Data   []string `json:"data"`
	Status string   `json:"status"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Datasource struct {
	UID string `json:"UID"`
}
