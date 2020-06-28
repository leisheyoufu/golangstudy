fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

.PHONY: fmt