scrapefd: *.go
	go vet
	golint
	go build

clean:
	go clean
	rm -f new.csv okcfd.csv scrapefd*
