package util

import (
	"awesomeProject/src/app/domain"
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const slugAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomSlug(n int) (string, error) {
	slug := make([]byte, n)
	for i := range slug {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugAlphabet))))
		if err != nil {
			return "", err
		}
		slug[i] = slugAlphabet[num.Int64()]
	}
	return string(slug), nil
}

func ParsePagination(c *gin.Context) domain.Pagination {
	var p domain.Pagination

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			p.Limit = uint(n)
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			p.Offset = uint(n)
		}
	}

	p.Normalize()
	return p
}

func HttpResponseFromError(err error) (int, map[string]any) {
	var code int

	switch {
	case errors.Is(err, domain.ErrVideoNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyArchived):
		code = http.StatusBadRequest
	case errors.Is(err, domain.ErrIncorrectUuid):
		code = http.StatusBadRequest
	default:
		code = http.StatusInternalServerError
	}

	return code, map[string]any{"error": err.Error()}
}
