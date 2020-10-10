package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"golang.org/x/xerrors"
)

type cognitoUser struct {
	// ユーザー名 = 識別情報
	username string
	// メールアドレス
	email string
	// 表示名
	name string
}

func newCongnitoUserFromUsername(username string) (cognitoUser, error) {
	idp := cognitoidentityprovider.New(awsSession)

	output, err := idp.AdminGetUser(&cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(os.Getenv("WISDOM_COGNITO_USER_POOL_ID")),
		Username:   aws.String(username),
	})
	if err != nil {
		return cognitoUser{}, xerrors.Errorf("get cognito user: %v", err)
	}

	// extract user attributes
	var email string
	var name string
	for _, attr := range output.UserAttributes {
		if *attr.Name == "email" {
			email = *attr.Value
		}
		if *attr.Name == "name" {
			name = *attr.Value
		}
	}

	return cognitoUser{
		username: username,
		email:    email,
		name:     name,
	}, nil
}
