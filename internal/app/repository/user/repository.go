package user

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/migdress/lambdaauthsamesls/internal/app/model"
	"github.com/pkg/errors"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	tableName      string
	dynamodbClient *dynamodb.DynamoDB
}

func NewUserRepository(
	tableName string,
	dynamodbClient *dynamodb.DynamoDB,
) *UserRepository {
	return &UserRepository{
		tableName,
		dynamodbClient,
	}
}

func (ur *UserRepository) hydrateUser(dynamodbRecord map[string]*dynamodb.AttributeValue) *model.User {
	user := model.User{}

	if v, ok := dynamodbRecord["email"]; ok {
		user.Email = *v.S
	}

	if v, ok := dynamodbRecord["password"]; ok {
		user.Password = *v.S
	}

	return &user
}
