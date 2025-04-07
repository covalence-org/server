package request

import (
	"covalence/src/types"
	"covalence/src/user"
	"errors"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelInfo stores information about a registered model
type rawRegister struct {
	Name     string  `json:"name" binding:"required"`
	Model    string  `json:"model" binding:"required"`
	APIURL   string  `json:"api_url" binding:"required"`
	Provider string  `json:"provider" binding:"required"`
	Status   *string `json:"status"`
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

	provider, err := types.NewModelProvider(r.Provider)
	if err != nil {
		return user.Model{}, errors.New("invalid model provider")
	}

	modelID, err := types.NewModelID(r.Model)
	if err != nil {
		return user.Model{}, errors.New("invalid model")
	}

	var status types.Status
	if r.Status != nil {
		status, err = types.NewStatus(*r.Status)
		if err != nil {
			return user.Model{}, errors.New("invalid status")
		}
	} else {
		status = types.Active()
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
		Provider:  provider,
		Status:    status,
	}, nil

}
