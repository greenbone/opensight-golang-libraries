// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package configReader

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOKReadEnvVarsIntoStruct(t *testing.T) {
	// given:
	type Test struct {
		Field1 string `viperEnv:"THE_VAR_1" default:"a string"`
		Field2 int    `viperEnv:"THE_VAR_2"`
		Nested struct {
			Field3 bool          `viperEnv:"THE_VAR_3"`
			Field4 string        `viperEnv:"THE_VAR_4" default:"a tttt"`
			Filed5 time.Duration `viperEnv:"THE_VAR_5" default:"1s"`
			Filed6 []int         `viperEnv:"THE_VAR_6" default:"14,30,60,90,180,365"`
		}
	}

	err := os.Setenv("THE_VAR_2", "42")
	assert.NoError(t, err)

	err = os.Setenv("THE_VAR_4", "Stringdafsf")
	assert.NoError(t, err)

	err = os.Setenv("THE_VAR_3", "true")
	assert.NoError(t, err)

	// when:
	var test Test
	_, err = ReadEnvVarsIntoStruct(&test)

	// then
	assert.NoError(t, err)
	assert.Equal(t, "a string", test.Field1)
	assert.Equal(t, 42, test.Field2)
	assert.Equal(t, true, test.Nested.Field3)
	assert.Equal(t, "Stringdafsf", test.Nested.Field4)
	assert.Equal(t, 1*time.Second, test.Nested.Filed5)
	assert.ElementsMatch(t, []int{14, 30, 60, 90, 180, 365}, test.Nested.Filed6)
}
