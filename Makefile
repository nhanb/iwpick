iwpick: main.go go.mod go.sum
	go build -o iwpick

run:
	go run *.go

watch:
	find . -name '*.go' | entr make run

install:
	go build -o ~/binaries/iwpick
