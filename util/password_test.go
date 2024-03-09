package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	password := RandomString(6)

	hpassword, err := HashedPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hpassword)

	err = CheckPassword(password, hpassword)
	require.NoError(t, err)

	wrongPassword := RandomString(6)

	err = CheckPassword(wrongPassword, hpassword)
	require.Error(t, err)
}
