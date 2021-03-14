package interfaces

import "github.com/aanufriev/httpproxy/internal/pkg/models"

type Usecase interface {
	SaveRequest(req models.Request) error
	GetRequests() ([]models.Request, error)
	GetRequest(id int) (models.Request, error)
}
