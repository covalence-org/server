package request

import (
	"errors"
	"netrunner/src/types"
	"netrunner/src/user"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelInfo stores information about a registered model
type RegisterRequest struct {
	Name   string `json:"name" binding:"required"`
	Model  string `json:"model" binding:"required"`
	ApiUrl string `json:"api_url" binding:"required"`
}

func ParseRegisterRequest(raw *gin.Context) (user.Model, error) {

	var payload RegisterRequest
	if err := raw.ShouldBindJSON(&payload); err != nil {
		return user.Model{}, errors.New("invalid request format")
	}

	name, err := types.NewName(payload.Name)
	if err != nil {
		return user.Model{}, errors.New("invalid name")
	}

	modelName, err := types.NewUserModel(payload.Model)
	if err != nil {
		return user.Model{}, errors.New("invalid model")
	}

	apiUrl, err := types.NewApiUrl(payload.ApiUrl)
	if err != nil {
		return user.Model{}, errors.New("invalid api url")
	}

	return user.Model{
		Name:      name,
		Model:     modelName,
		ApiUrl:    apiUrl,
		CreatedAt: time.Now(),
	}, nil

}
