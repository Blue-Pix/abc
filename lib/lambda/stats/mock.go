package stats

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/stretchr/testify/mock"
)

type MockLambdaClient struct {
	mock.Mock
	lambdaiface.LambdaAPI
}

func (client *MockLambdaClient) ListFunctions(params *lambda.ListFunctionsInput) (*lambda.ListFunctionsOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*lambda.ListFunctionsOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func SetMockDefaultBehaviour(lm *MockLambdaClient) {
	lm.On("ListFunctions", &lambda.ListFunctionsInput{
		Marker:   nil,
		MaxItems: aws.Int64(1000),
	}).Return(
		&lambda.ListFunctionsOutput{
			NextMarker: aws.String("next_marker"),
			Functions: []*lambda.FunctionConfiguration{
				{Runtime: aws.String("nodejs12.x"), FunctionName: aws.String("node12-func-1")},
				{Runtime: aws.String("nodejs10.x"), FunctionName: aws.String("node10-func-1")},
				{Runtime: aws.String("python3.8"), FunctionName: aws.String("python3.8-func-1")},
			},
		},
		nil,
	)
	lm.On("ListFunctions", &lambda.ListFunctionsInput{
		Marker:   aws.String("next_marker"),
		MaxItems: aws.Int64(1000),
	}).Return(
		&lambda.ListFunctionsOutput{
			NextMarker: nil,
			Functions: []*lambda.FunctionConfiguration{
				{Runtime: aws.String("python3.8"), FunctionName: aws.String("python3.8-func-2")},
				{Runtime: aws.String("python3.8"), FunctionName: aws.String("python3.8-func-3")},
				{Runtime: aws.String("python3.7"), FunctionName: aws.String("python3.7-func-1")},
				{Runtime: aws.String("python3.6"), FunctionName: aws.String("python3.6-func-1")},
				{Runtime: aws.String("ruby2.7"), FunctionName: aws.String("ruby2.7-func-1")},
				{Runtime: aws.String("ruby2.5"), FunctionName: aws.String("ruby2.5-func-1")},
				{Runtime: aws.String("java11"), FunctionName: aws.String("java11-func-1")},
				{Runtime: aws.String("java8"), FunctionName: aws.String("java8-func-1")},
				{Runtime: aws.String("go1.x"), FunctionName: aws.String("go1-func-1")},
				{Runtime: aws.String("go1.x"), FunctionName: aws.String("go1-func-2")},
				{Runtime: aws.String("go1.x"), FunctionName: aws.String("go1-func-3")},
				{Runtime: aws.String("go1.x"), FunctionName: aws.String("go1-func-4")},
				{Runtime: aws.String("dotnetcore3.1"), FunctionName: aws.String("dotnet3.1-func-1")},
				{Runtime: aws.String("dotnetcore2.1"), FunctionName: aws.String("dotnet2.1-func-1")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-1")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-2")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-3")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-4")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-5")},
				{Runtime: aws.String("provided"), FunctionName: aws.String("provided-func-6")},
				// deprecated
				{Runtime: aws.String("nodejs4.3"), FunctionName: aws.String("nodejs4.3-func-1")},
				{Runtime: aws.String("nodejs4.3-edge"), FunctionName: aws.String("nodejs4.3edge-func-1")},
				{Runtime: aws.String("nodejs6.10"), FunctionName: aws.String("nodejs6.10-func-1")},
				{Runtime: aws.String("nodejs8.10"), FunctionName: aws.String("nodejs8.10-func-1")},
				{Runtime: aws.String("nodejs"), FunctionName: aws.String("nodejs0.10-func-1")},
				{Runtime: aws.String("nodejs"), FunctionName: aws.String("nodejs0.10-func-2")},
				{Runtime: aws.String("dotnetcore1.0"), FunctionName: aws.String("dotnet1.0-func-1")},
				{Runtime: aws.String("dotnetcore2.0"), FunctionName: aws.String("dotnet2.0-func-1")},
			},
		},
		nil,
	)
}
