package athena

import "pvk/API/db"

type Athena struct {
	DBClient *db.DBClient
}

type Match struct {
	Payload struct {
		State struct {
			ID        int    `json:"id"`
			Lifecycle string `json:"lifecycle"`
			Order     int    `json:"order"`
			Series    struct {
				ID int `json:"id"`
			} `json:"series"`
			Map struct {
				ID int `json:"id"`
			} `json:"map"`
			DeletedAt interface{} `json:"deleted_at"`
			Game      struct {
				ID int `json:"id"`
			} `json:"game"`
			Participants []struct {
				Seed    int         `json:"seed"`
				Score   interface{} `json:"score"`
				Forfeit bool        `json:"forfeit"`
				Roster  struct {
					ID int `json:"id"`
				} `json:"roster"`
				Winner bool        `json:"winner"`
				Stats  interface{} `json:"stats"`
			} `json:"participants"`
		} `json:"state"`
	} `json:"payload"`
}

type Cacheable interface {
	Tournament | Game | Roster
}

type cache[T Cacheable] struct {
	prefix string
}

type Series struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	Start        string     `json:"start"`
	End          string     `json:"end"`
	Lifecycle    string     `json:"lifecycle"`
	Tier         int        `json:"tier"`
	BestOf       int        `json:"best_of"`
	Streamed     bool       `json:"streamed"`
	Tournament   Tournament `json:"tournament"`
	Game         Game       `json:"game"`
	Participants []struct {
		Seed   int    `json:"seed"`
		Winner bool   `json:"winner"`
		Roster Roster `json:"roster"`
	} `json:"participants"`
	Casters []struct {
		Primary bool `json:"primary"`
		Caster  struct {
			ID int `json:"id"`
		} `json:"caster"`
	} `json:"casters"`
	Broadcasters []struct {
		Official    bool `json:"official"`
		Broadcaster struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Platform struct {
				ID int `json:"id"`
			} `json:"platform"`
		} `json:"broadcaster"`
	} `json:"broadcasters"`
}

type Tournament struct {
	ID         int         `json:"id"`
	Title      string      `json:"title"`
	ShortTitle string      `json:"short_title"`
	Tier       int         `json:"tier"`
	Start      string      `json:"start"`
	End        string      `json:"end"`
	DeletedAt  interface{} `json:"deleted_at"`
}

type Roster struct {
	ID     int  `json:"id"`
	Team   Team `json:"team"`
	LineUp struct {
		ID      int `json:"id"`
		Players []struct {
			ID int `json:"id"`
		} `json:"players"`
	} `json:"line_up"`
	Game struct {
		ID int `json:"id"`
	} `json:"game"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Game struct {
	ID           int    `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Title        string `json:"title"`
	Images       []struct {
		URL string `json:"url"`
	} `json:"images"`
}
