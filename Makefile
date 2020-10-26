# stevebot makefile
SRC=main.go internal/steve/*.go internal/bot/*.go internal/sys/*.go

all: bin/stevebot

bin/stevebot: $(SRC)
	mkdir -p bin/
	go build -o bin/stevebot -mod vendor

clean:
	rm -r bin
.PHONY: clean
