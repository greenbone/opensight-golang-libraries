// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/configReader"
	"github.com/rs/zerolog/log"
	"time"
)

type OpensearchClientConfig struct {
	Host                 string        `validate:"required" viperEnv:"ELASTIC_HOST"`
	Port                 int           `validate:"required,min=1,max=65535" viperEnv:"ELASTIC_API_PORT"`
	Https                bool          `viperEnv:"ELASTIC_HTTPS"`
	Username             string        `viperEnv:"ELASTIC_USER"`
	Password             string        `viperEnv:"ELASTIC_PASS"`
	UpdateMaxRetries     int           `validate:"required" viperEnv:"OPEN_SEARCH_UPDATE_MAX_RETRIES" default:"10"`
	UpdateRetrySleep     time.Duration `validate:"required" viperEnv:"OPEN_SEARCH_UPDATE_RETRY_SLEEP" default:"2s"`
	KeycloakClient       string        `viperEnv:"OPENSEARCH_KEYCLOAK_CLIENT" default:"opensearch-client"`
	KeycloakClientSecret string        `viperEnv:"OPENSEARCH_KEYCLOAK_CLIENT_SECRET"`
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
