.PHONY: docker run container

CONTAINERS := container-test
BIN := container-test
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go: $(BIN)

clean:
	rm -f $(BIN)

container: go
	docker image build -t container-test .

container-test: container-test.go
	go build $^

test-codebuild:
	./codebuild_build.sh -i public.ecr.aws/codebuild/amazonlinux2-x86_64-standard:4.0 -a outdir

# run: docker
# 	docker run --rm -it --name sample01 -p 8080:8080 sample
