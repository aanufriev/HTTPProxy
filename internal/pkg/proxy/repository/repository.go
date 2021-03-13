package repository

import (
	"database/sql"

	"github.com/aanufriev/httpproxy/internal/pkg/models"
	"github.com/aanufriev/httpproxy/internal/pkg/proxy/interfaces"
)

type ProxyRepository struct {
	db *sql.DB
}

func NewProxyRepository(db *sql.DB) interfaces.Repository {
	return ProxyRepository{
		db: db,
	}
}

func (r ProxyRepository) SaveRequest(req models.Request) error {
	_, err := r.db.Exec(
		`INSERT INTO requests (method, host, scheme, path, headers, body)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.Method, req.Host, req.Scheme, req.Path, req.Headers, req.Body,
	)

	return err
}

func (r ProxyRepository) GetRequests() ([]models.Request, error) {
	rows, err := r.db.Query("SELECT id, method, host, scheme, path, headers, body FROM requests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]models.Request, 0)
	req := models.Request{}
	for rows.Next() {
		err = rows.Scan(
			&req.ID, &req.Method, &req.Host, &req.Scheme,
			&req.Path, &req.Headers, &req.Body,
		)

		if err != nil {
			return nil, err
		}
	}

	return requests, nil
}