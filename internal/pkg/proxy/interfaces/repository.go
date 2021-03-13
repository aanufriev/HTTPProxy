package interfaces

import "github.com/aanufriev/httpproxy/internal/pkg/models"

type Repository interface {
	SaveRequest(req models.Request) error
	GetRequests() ([]models.Request, error)
}
