// Copyright (C) Greenbone Networks GmbH
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/greenbone/opensight-golang-libraries/pkg/configReader"
	"github.com/rs/zerolog/log"
)

// CryptoConfig defines the configuration for service-wide cryptography options.
//
// Version specific options will apply to all newer encryption versions up to the variants with a new specific option,
// e.g. if the options MyKeyV1 and MyKeyV4 exist, MyKeyV1 will apply to v1, v2 and v3, while MyKeyV4 applies to v4 and
// newer.
type CryptoConfig struct {
	// Contains the password for encrypting user group specific report encryptions using v1 to v2
	ReportEncryptionV1Password string `validate:"required" viperEnv:"TASK_REPORT_CRYPTO_V1_PASSWORD"`
	// Contains the salt for encrypting user group specific report encryptions v1 to v2
	ReportEncryptionV1Salt string `validate:"required,gte=32" viperEnv:"TASK_REPORT_CRYPTO_V1_SALT"`
}

func Read() (config CryptoConfig, err error) {
	_, err = configReader.ReadEnvVarsIntoStruct(&config)
	if err != nil {
		return config, err
	}
	log.Debug().Msgf("CryptoConfig: %+v", config)
	return config, nil
}
