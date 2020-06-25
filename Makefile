.PHONY: docker run container

SERVERS := info server_p1 server_c1 server_c2
GOOS=linux
GOARCH=amd64

go: $(SERVERS)

clean:
	rm -f $(SERVERS)

container: go $(foreach t,$(PARTS),container/$(t))

container/%: Dockerfiles/%
	docker image build -t $(patsubst Dockerfiles/%,%,$<) -f $< .

$(SERVERS): %: %.go server_lib.go
	go build $^

# run: docker
# 	docker run --rm -it --name sample01 -p 8080:8080 sample
