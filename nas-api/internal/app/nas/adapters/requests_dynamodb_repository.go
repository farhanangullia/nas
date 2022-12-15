package adapters

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"nas/internal/app/nas"
	"nas/internal/app/nas/common"
	"log"
	"time"
)

type RequestsDynamoDbRepository struct {
	db *dynamodb.Client
}

// returns a concrete repository backed by DynamoDB
func NewRequestsDynamoDbRepository(db *dynamodb.Client) nas.RequestsRepository {
	if db == nil {
		panic("missing db")
	}

	// Creates a DynamoDB table with a primary key defined as
	// a key named `id`.
	// Uses NewTableExistsWaiter to wait for the table to be created by
	// DynamoDB before it returns.

	tableName := "IpAllowListRequests" // TODO: Config this ext
	exists, err := common.TableExists(context.TODO(), tableName, db)
	if err != nil {
		panic(err)
	}
	if !exists {
		log.Printf("Creating table %v...\n", tableName)
		_, err := db.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{{
				AttributeName: aws.String("Id"),
				AttributeType: types.ScalarAttributeTypeS,
			}},
			KeySchema: []types.KeySchemaElement{{
				AttributeName: aws.String("Id"),
				KeyType:       types.KeyTypeHash,
			}},
			TableName: aws.String(tableName),
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(10),
				WriteCapacityUnits: aws.Int64(10),
			},
		})

		if err != nil {
			log.Printf("Couldn't create table %v. Here's why: %v\n", tableName, err)
		} else {
			waiter := dynamodb.NewTableExistsWaiter(db)
			err = waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName)}, 5*time.Minute)
			if err != nil {
				log.Printf("Wait for table exists failed. Here's why: %v\n", err)
			} else {
				log.Printf("Created table: %v\n", tableName)
			}
		}
	} else {
		log.Printf("Table %v already exists.\n", tableName)
	}

	return &RequestsDynamoDbRepository{
		db: db,
	}
}

func (r *RequestsDynamoDbRepository) CreateIpAccessRequest(ctx context.Context, request nas.IpAccessRequest) error {

	item, err := attributevalue.MarshalMap(request)
	if err != nil {
		panic(err)
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("IpAllowListRequests"), Item: item,
	})
	if err != nil {
		log.Printf("Couldn't add item to table. Here's why: %v\n", err)
		return err
	}

	return nil
}

func (r *RequestsDynamoDbRepository) RetrieveIpAccessRequest(ctx context.Context, requestId string) (*nas.IpAccessRequest, error) {
	ipAccessRequest := nas.IpAccessRequest{}
	reqId, err := attributevalue.Marshal(requestId)
	if err != nil {
		panic(err)
	}

	key := map[string]types.AttributeValue{"Id": reqId}
	//key := map[string]types.AttributeValue{"Id": &types.AttributeValueMemberS{Value: requestId}}

	response, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		Key: key, TableName: aws.String("IpAllowListRequests"),
	})
	if err != nil {
		log.Printf("Couldn't get info about %v. Here's why: %v\n", key, err)
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &ipAccessRequest)
		if err != nil {
			log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
		}
	}
	return &ipAccessRequest, err
}
