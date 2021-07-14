package user

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/lambda-custom-authorizers/lambdaauth/internal/app/model"
	"github.com/pkg/errors"
)

func (ur *UserRepository) Save(userModel model.User) error {
	_, err := ur.dynamodbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ur.tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(userModel.Email),
			},
			"password": {
				S: aws.String(userModel.Password),
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "userrepository: UserRepository.Save dynamodbClient.PutItem error")
	}

	return nil
}
