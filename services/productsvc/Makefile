.PHONY: all
all:
	migrations
	graph

.PHONY: migrations
migrations:

.PHONY: graph
graph:
	go run github.com/99designs/gqlgen generate

.PHONY: clean
clean:
	$(RM) graph/generated.go
	$(RM) graph/model/models_gen.go
