package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	OWM_APIURL = "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key"
)

func TestOWMSegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		ExpectedEnabled bool
		Error           error
	}{
		{
			Case:            "Sunny Display",
			JSONResponse:    `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:  "\ufa98 (20°C)",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error in retrieving data",
			JSONResponse:    "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				APIKEY:   "key",
				LOCATION: "AMSTERDAM,NL",
				UNITS:    "metric",
			},
		}

		env.On("doGet", OWM_APIURL).Return([]byte(tc.JSONResponse), tc.Error)

		o := &owm{
			props: props,
			env:   env,
		}

		enabled := o.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, o.string(), tc.Case)
	}
}

func TestOWMSegmentIcons(t *testing.T) {
	cases := []struct {
		Case               string
		IconID             string
		ExpectedIconString string
	}{
		{
			Case:               "Sunny Display day",
			IconID:             "01d",
			ExpectedIconString: "\ufa98",
		},
		{
			Case:               "Light clouds Display day",
			IconID:             "02d",
			ExpectedIconString: "\ufa94",
		},
		{
			Case:               "Cloudy Display day",
			IconID:             "03d",
			ExpectedIconString: "\ue33d",
		},
		{
			Case:               "Broken Clouds Display day",
			IconID:             "04d",
			ExpectedIconString: "\ue312",
		},
		{
			Case:               "Shower Rain Display day",
			IconID:             "09d",
			ExpectedIconString: "\ufa95",
		},
		{
			Case:               "Rain Display day",
			IconID:             "10d",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display day",
			IconID:             "11d",
			ExpectedIconString: "\ue31d",
		},
		{
			Case:               "Snow Display day",
			IconID:             "13d",
			ExpectedIconString: "\ue31a",
		},
		{
			Case:               "Fog Display day",
			IconID:             "50d",
			ExpectedIconString: "\ue313",
		},

		{
			Case:               "Sunny Display night",
			IconID:             "01n",
			ExpectedIconString: "\ufa98",
		},
		{
			Case:               "Light clouds Display night",
			IconID:             "02n",
			ExpectedIconString: "\ufa94",
		},
		{
			Case:               "Cloudy Display night",
			IconID:             "03n",
			ExpectedIconString: "\ue33d",
		},
		{
			Case:               "Broken Clouds Display night",
			IconID:             "04n",
			ExpectedIconString: "\ue312",
		},
		{
			Case:               "Shower Rain Display night",
			IconID:             "09n",
			ExpectedIconString: "\ufa95",
		},
		{
			Case:               "Rain Display night",
			IconID:             "10n",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display night",
			IconID:             "11n",
			ExpectedIconString: "\ue31d",
		},
		{
			Case:               "Snow Display night",
			IconID:             "13n",
			ExpectedIconString: "\ue31a",
		},
		{
			Case:               "Fog Display night",
			IconID:             "50n",
			ExpectedIconString: "\ue313",
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				APIKEY:   "key",
				LOCATION: "AMSTERDAM,NL",
				UNITS:    "metric",
			},
		}

		response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, tc.IconID)
		expectedString := fmt.Sprintf("%s (20°C)", tc.ExpectedIconString)

		env.On("doGet", OWM_APIURL).Return([]byte(response), nil)

		o := &owm{
			props: props,
			env:   env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, o.string(), tc.Case)
	}
}
func TestOWMSegmentValidCachePathAndNotTimedout(t *testing.T) {
	response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, "01d")
	expectedString := fmt.Sprintf("%s (20°C)", "\ufa98")

	// create a fake cache file
	f, err := os.CreateTemp("", "")
	assert.Equal(t, nil, err)
	_, err = f.Write([]byte(response))
	assert.Equal(t, nil, err)
	f.Close()
	defer os.Remove(f.Name()) // clean up

	env := &MockedEnvironment{}
	props := &properties{
		values: map[Property]interface{}{
			APIKEY:    "key",
			LOCATION:  "AMSTERDAM,NL",
			UNITS:     "metric",
			CACHEFILE: filepath.Base(f.Name()),
		},
	}
	o := &owm{
		props: props,
		env:   env,
	}
	env.On("doGet", OWM_APIURL).Return(nil, nil)

	assert.Nil(t, o.setStatus())
	assert.Equal(t, expectedString, o.string())
}

func TestOWMSegmentValidCachePathAndTimedout(t *testing.T) {
	cacheData := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, "01d")
	response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":22}}`, "01d")
	expectedString := fmt.Sprintf("%s (20°C)", "\ufa98")

	// create a fake cache file
	f, err := os.CreateTemp("", "")
	assert.Equal(t, nil, err)
	_, err = f.Write([]byte(cacheData))
	assert.Equal(t, nil, err)
	f.Close()
	defer os.Remove(f.Name()) // clean up

	env := &MockedEnvironment{}
	props := &properties{
		values: map[Property]interface{}{
			APIKEY:    "key",
			LOCATION:  "AMSTERDAM,NL",
			UNITS:     "metric",
			CACHEFILE: filepath.Base(f.Name()),
		},
	}
	o := &owm{
		props: props,
		env:   env,
	}
	env.On("doGet", OWM_APIURL).Return([]byte(response), nil)

	assert.Nil(t, o.setStatus())
	assert.Equal(t, expectedString, o.string())
}

func TestOWMSegmentValidCachePathWrongData(t *testing.T) {
	cacheData := "hello world"

	// create a fake cache file
	f, err := os.CreateTemp("", "")
	assert.Equal(t, nil, err)
	_, err = f.Write([]byte(cacheData))
	assert.Equal(t, nil, err)
	f.Close()
	defer os.Remove(f.Name()) // clean up

	env := &MockedEnvironment{}
	props := &properties{
		values: map[Property]interface{}{
			APIKEY:    "key",
			LOCATION:  "AMSTERDAM,NL",
			UNITS:     "metric",
			CACHEFILE: filepath.Base(f.Name()),
		},
	}
	o := &owm{
		props: props,
		env:   env,
	}

	err = o.setStatus()
	assert.EqualError(t, err, "invalid character 'h' looking for beginning of value")
}

func TestOWMSegmentInvalidCachePath(t *testing.T) {
	env := &MockedEnvironment{}
	props := &properties{
		values: map[Property]interface{}{
			APIKEY:    "key",
			LOCATION:  "AMSTERDAM,NL",
			UNITS:     "metric",
			CACHEFILE: "*[]\\/",
		},
	}
	o := &owm{
		props: props,
		env:   env,
	}

	expectedErrorMessage := fmt.Sprintf("CreateFile %s\\*[]: The filename, directory name, or volume label syntax is incorrect.", os.TempDir())

	err := o.setStatus()
	assert.EqualError(t, err, expectedErrorMessage)
}
