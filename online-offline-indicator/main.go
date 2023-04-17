package online_offline_indicator

import (
	"online_offline_indicator/service"

	"github.com/rs/zerolog/log"
)

func GetService() *service.Service {
	svc, err := service.NewService()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create service")
	}
	return svc
}
