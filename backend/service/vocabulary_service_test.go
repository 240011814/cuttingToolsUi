package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestVocabularyService(t *testing.T) {
	svc := NewVocabularyService()
	assert.NotNil(t, svc)
}
