.PHONY: docker run

SERVERS := server_p1 server_c1 server_c2
GOOS=linux
GOARCH=amd64

go: info $(SERVERS)

container: go $(foreach t,$(PARTS),container-$(t))

container-%: Dockerfile.%
	docker image build -t $(patsubst Dockerfile.%,%,$<) -f $< .

$(SERVERS): %: %.go server_lib.go
	go build $^

%: %.go
	go build $<

# run: docker
# 	docker run --rm -it --name sample01 -p 8080:8080 sample
