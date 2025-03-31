package request

import (
	"errors"
	"net/url"
	"netrunner/src/types"
	"netrunner/src/user"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelInfo stores information about a registered model
type rawRegister struct {
	Name   string `json:"name" binding:"required"`
	Model  string `json:"model" binding:"required"`
	APIURL string `json:"api_url" binding:"required"`
}

func ParseRegister(c *gin.Context) (user.Model, error) {

	var r rawRegister
	if err := c.ShouldBindJSON(&r); err != nil {
		return user.Model{}, err
	}

	name, err := types.NewName(r.Name)
	if err != nil {
		return user.Model{}, errors.New("invalid name")
	}

	modelID, err := types.NewModelID(r.Model)
	if err != nil {
		return user.Model{}, errors.New("invalid model")
	}

	// Build target URL
	apiURL, err := url.Parse(r.APIURL)
	if err != nil {
		return user.Model{}, errors.New("invalid api url")
	}

	return user.Model{
		Name:      name,
		Model:     modelID,
		APIURL:    apiURL,
		CreatedAt: time.Now(),
	}, nil

}
