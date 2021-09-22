package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemory(t *testing.T) {
	cache := New(5 * time.Minute)

	cache.put("my_key", "xxx", DefaultExpiration)
	result, found := cache.get("my_key")

	assert.Equal(t, result, "xxx")
	assert.Equal(t, found, true)
}

func TestSaveToFile(t *testing.T) {
	cache := New(5 * time.Minute)

	cache.put("my_key", "xxx", DefaultExpiration)
	result, found := cache.get("my_key")

	assert.Equal(t, result, "xxx")
	assert.Equal(t, found, true)

	file, err := ioutil.TempFile(os.TempDir(), "omp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	err = cache.saveToFile(file.Name())
	assert.Equal(t, err, nil)
}

func TestLoadFromFileExpiredKey(t *testing.T) {
	cache := New(5 * time.Minute)

	_ = cache.loadFromFile("omp431810701")

	result, found := cache.get("my_key")

	assert.Equal(t, result, "xxx")
	assert.Equal(t, found, true)
}
