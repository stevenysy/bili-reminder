FUNCTION_NAME=bili-reminder
ZIP_PATH=bili-reminder.zip

all: | build zip deploy

build:
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o bootstrap

zip:
	zip $(ZIP_PATH) bootstrap

deploy:
	aws lambda update-function-code --function-name $(FUNCTION_NAME) --zip-file fileb://$(ZIP_PATH) > /dev/null