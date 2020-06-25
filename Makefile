.PHONY: docker run container

CONTAINERS := server_p1 server_c1 server_c2
BIN := info $(CONTAINERS)
GOOS=linux
GOARCH=amd64

go: $(BIN)

clean:
	rm -f $(BIN)

container: go $(foreach t,$(CONTAINERS),container/$(t))

container/%: Dockerfiles/%
	docker image build -t $(patsubst Dockerfiles/%,%,$<) -f $< .

$(BIN): %: src/%.go src/server_lib.go
	go build $^

# run: docker
# 	docker run --rm -it --name sample01 -p 8080:8080 sample
