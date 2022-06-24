package main

import "github.com/rs/zerolog/log"

func main() {
	config, err := ReadConfig()
	if err != nil {
		log.Panic().Err(err).Msgf("there was problem when reading config, stop execution")
	}

	err = writeMessage(config)
	if err != nil {
		log.Panic().Err(err).Msgf("there was problem when writing message, stop execution")
	}

	log.Info().Msg("finished")
}
