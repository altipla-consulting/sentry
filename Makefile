
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

	go get -u github.com/getsentry/raven-go
	go get -u github.com/juju/errors
	go get -u github.com/sirupsen/logrus
	go get -u golang.org/x/net/context

test:
	revive -formatter friendly
	go install .
