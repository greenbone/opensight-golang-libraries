// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/configReader"
	"github.com/rs/zerolog/log"
)

type OpensearchClientConfig struct {
	Host         string `validate:"required" viperEnv:"ELASTIC_HOST"`
	Port         int    `validate:"required,min=1,max=65535" viperEnv:"ELASTIC_API_PORT"`
	Https        bool   `viperEnv:"ELASTIC_HTTPS"`
	AuthUsername string `validate:"required" viperEnv:"ELASTIC_AUTH_USER"`
	AuthPassword string `validate:"required" viperEnv:"ELASTIC_AUTH_PASS"`
	AuthMethod   string `validate:"required" viperEnv:"ELASTIC_AUTH_METHOD"`
}

func ReadOpensearchClientConfig() (OpensearchClientConfig, error) {
	config := &OpensearchClientConfig{}
	_, err := configReader.ReadEnvVarsIntoStruct(config)
	if err != nil {
		return *config, err
	}
	log.Debug().Msgf("OpenSearch Config: %+v", config)
	return *config, nil
}
