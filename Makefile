##################################################
# Golang
##################################################
.PHONY: install build

OUTPUT_PATH = ./app

install:
	go install github.com/aws/aws-lambda-go/lambda@latest

build:
	go build -o ${OUTPUT_PATH} ./main.go

##################################################
# AWS - Amazon ECR
##################################################
.PHONY: auth-registry create-repository build-image push-image

REGISTRY_ID     = $(shell aws ecr describe-registry | jq -r .registryId)
REPOSITORY_NAME = $(shell aws ecr describe-repositories | jq -r '.repositories[0].repositoryName')
AWS_REGION      = ap-northeast-1
IMAGE_TAG       = latest

auth-registry:
	aws ecr get-login-password --region ${AWS_REGION} \
	| docker login --username AWS --password-stdin ${REGISTRY_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com

create-repository:
	aws ecr create-repository --repository-name ${REPOSITORY_NAME} --region ${AWS_REGION}

build-image:
	docker build -t ${REPOSITORY_NAME}:${IMAGE_TAG} .

push-image:
	docker tag ${REPOSITORY_NAME}:${IMAGE_TAG} ${REGISTRY_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${REPOSITORY_NAME}:latest
	docker push ${REGISTRY_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com/${REPOSITORY_NAME}:latest
