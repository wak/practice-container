.PHONY: docker run

PARTS := info server_p1 server_c1 server_c2


go: $(PARTS)

container: go $(foreach t,$(PARTS),container-$(t))

container-%: Dockerfile.%
	docker image build -t $(patsubst Dockerfile.%,%,$<) -f $< .

%: %.go
	GOOS=linux GOARCH=amd64 go build $<

# run: docker
# 	docker run --rm -it --name sample01 -p 8080:8080 sample
