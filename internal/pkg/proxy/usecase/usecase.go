package usecase

import (
	"github.com/aanufriev/httpproxy/internal/pkg/models"
	"github.com/aanufriev/httpproxy/internal/pkg/proxy/interfaces"
)

type ProxyUsecase struct {
	proxyRepository interfaces.Repository
}

func NewProxyUsecase(proxyRepository interfaces.Repository) ProxyUsecase {
	return ProxyUsecase{
		proxyRepository: proxyRepository,
	}
}

func (u ProxyUsecase) SaveRequest(req models.Request) error {
	return u.proxyRepository.SaveRequest(req)
}

func (u ProxyUsecase) GetRequests() ([]models.Request, error) {
	return u.proxyRepository.GetRequests()
}
