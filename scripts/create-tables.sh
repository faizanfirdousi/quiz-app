#!/bin/bash
# Creates all DynamoDB tables for KahootClone on DynamoDB Local.
# Prerequisites: DynamoDB Local running on port 8000
#   docker run -d -p 8000:8000 amazon/dynamodb-local
set -e

export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_DEFAULT_REGION=ap-south-1

ENDPOINT="--endpoint-url http://localhost:8000"

echo "Creating kahootclone-quizzes table..."
aws dynamodb create-table \
  --table-name kahootclone-quizzes \
  --attribute-definitions AttributeName=quizId,AttributeType=S \
  --key-schema AttributeName=quizId,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  $ENDPOINT

echo "Creating kahootclone-sessions table..."
aws dynamodb create-table \
  --table-name kahootclone-sessions \
  --attribute-definitions \
    AttributeName=sessionId,AttributeType=S \
    AttributeName=pin,AttributeType=S \
  --key-schema AttributeName=sessionId,KeyType=HASH \
  --global-secondary-indexes '[{
    "IndexName":"pin-index",
    "KeySchema":[{"AttributeName":"pin","KeyType":"HASH"}],
    "Projection":{"ProjectionType":"ALL"}
  }]' \
  --billing-mode PAY_PER_REQUEST \
  $ENDPOINT

echo "Creating kahootclone-connections table..."
aws dynamodb create-table \
  --table-name kahootclone-connections \
  --attribute-definitions \
    AttributeName=sessionId,AttributeType=S \
    AttributeName=connectionId,AttributeType=S \
  --key-schema AttributeName=sessionId,KeyType=HASH AttributeName=connectionId,KeyType=RANGE \
  --global-secondary-indexes '[{
    "IndexName":"connectionId-index",
    "KeySchema":[{"AttributeName":"connectionId","KeyType":"HASH"}],
    "Projection":{"ProjectionType":"ALL"}
  }]' \
  --billing-mode PAY_PER_REQUEST \
  $ENDPOINT

echo "Creating kahootclone-answers table..."
aws dynamodb create-table \
  --table-name kahootclone-answers \
  --attribute-definitions \
    AttributeName=sessionId,AttributeType=S \
    AttributeName=userIdQuestionId,AttributeType=S \
  --key-schema AttributeName=sessionId,KeyType=HASH AttributeName=userIdQuestionId,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  $ENDPOINT

echo ""
echo "All tables created. Verifying..."
aws dynamodb list-tables $ENDPOINT

echo "Done!"
