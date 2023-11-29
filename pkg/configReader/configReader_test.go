package configReader

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOKReadEnvVarsIntoStruct(t *testing.T) {
	type Test struct {
		Field1 string `viperEnv:"THE_VAR_1" default:"a string"`
		Field2 int    `viperEnv:"THE_VAR_2"`
		Nested struct {
			Field3 bool          `viperEnv:"THE_VAR_3"`
			Field4 string        `viperEnv:"THE_VAR_4" default:"a tttt"`
			Filed5 time.Duration `viperEnv:"THE_VAR_5" default:"1s"`
		}
	}

	err := os.Setenv("THE_VAR_2", "42")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = os.Setenv("THE_VAR_4", "Stringdafsf")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = os.Setenv("THE_VAR_3", "true")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	var test Test
	_, err = ReadEnvVarsIntoStruct(&test)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "a string", test.Field1)
	assert.Equal(t, 42, test.Field2)
	assert.Equal(t, true, test.Nested.Field3)
	assert.Equal(t, "Stringdafsf", test.Nested.Field4)
	assert.Equal(t, time.Second, test.Nested.Filed5)
}
