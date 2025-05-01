package repo

import (
	"cloud/bucket"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ID	TokenAmount	 Rate(time of additing new tokens)

type Repo struct {
	db *sql.DB
}

func New(port string, username string, host string, DBname string, password string) *Repo {
	// dsn := fmt.Sprintf("postgres://postgres:%s@%s:%s/%s", password, host, port, DBname)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, DBname)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("DB connection is failed:%v", err)
	}
	err = db.Ping()
	if err != nil {
		log.Printf("bad connection:%v", err)
	}

	return &Repo{db: db}
}

func (r *Repo) GetID(id string) (*bucket.Client, error) {
	stmt, err := r.db.Prepare(`SELECT ID, Rate, Capacity FROM Storage WHERE Id = $1`)
	if err != nil {
		return nil, fmt.Errorf("Prepare info error:%v", err)
	}
	defer stmt.Close()

	res := stmt.QueryRow(id)

	p := bucket.Client{}
	err = res.Scan(&p.ID, &p.Rate, &p.Capacity)
	if err != nil {
		return nil, fmt.Errorf("failed to scan:%v", err)
	}
	return &p, nil

}
func (r *Repo) GetALL() ([]*bucket.Client, error) {
	query := `SELECT Id, Item, Quantity FROM Storage`

	var clients []*bucket.Client
	stmt, err := r.db.Prepare(query)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer stmt.Close()

	res, err := stmt.Query()
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer res.Close()

	for res.Next() {
		var p bucket.Client
		err := res.Scan(&p.ID, &p.Capacity, &p.Rate)
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		clients = append(clients, &p)
	}

	if err = res.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %v", err)
	}
	return clients, nil
}

func (r *Repo) AddCLient(c *bucket.Client) error {
	stmt, err := r.db.Prepare(`INSERT INTO Storage (Id, Rate, Capacity) VALUES ($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("Prepare info error:%v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.ID, c.Rate, c.Capacity)
	if err != nil {
		return fmt.Errorf("setting info error:%v", err)

	}
	return nil
}
