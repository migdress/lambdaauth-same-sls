package main

import (
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/migdress/lambdaauthsamesls/internal/jwtwrapper"
	"github.com/pkg/errors"
)

type Handler struct {
	jwtWrapper jwtWrapper
}

func NewHandler(
	jwtWrapper jwtWrapper,
) *Handler {
	return &Handler{
		jwtWrapper,
	}
}

type jwtWrapper interface {
	VerifyToken(tokenString string) (map[string]string, error)
}

func (h *Handler) Invoke() func(request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return func(request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		token := request.AuthorizationToken
		tokenSlice := strings.Split(token, " ")

		if len(tokenSlice) == 0 {
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
		}

		bearerToken := tokenSlice[len(tokenSlice)-1]
		claims, err := h.jwtWrapper.VerifyToken(bearerToken)
		if err != nil {
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
		}
		/*
			if bearerToken != "hello" {
				return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
			}
		*/

		return h.generatePolicy(claims["email"], "Allow", request.MethodArn), nil
	}
}

func (h *Handler) generatePolicy(principalID, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
	}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	return authResponse
}

func main() {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		panic("SECRET_KEY cannot be empty")
	}

	jwtWrapper := jwtwrapper.NewJWTWrapper(secretKey, time.Minute*15)

	handler := NewHandler(jwtWrapper)

	lambda.Start(handler.Invoke())
}
