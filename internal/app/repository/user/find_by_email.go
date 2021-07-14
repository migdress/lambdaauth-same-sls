package user

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/migdress/lambdaauthsamesls/internal/app/model"
	"github.com/pkg/errors"
)

func (ur *UserRepository) FindByEmail(email string) (*model.User, error) {
	res, err := ur.dynamodbClient.Query(&dynamodb.QueryInput{
		TableName: aws.String(ur.tableName),
		KeyConditions: map[string]*dynamodb.Condition{
			"email": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(email),
					},
				},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "repository: UserRepository.FindByEmail dynamodbClient.Query error")
	}

	if len(res.Items) == 0 {
		return nil, errors.Wrap(ErrUserNotFound, "repository: UserRepository.FindByEmail user not found")
	}

	return ur.hydrateUser(res.Items[0]), nil
}
