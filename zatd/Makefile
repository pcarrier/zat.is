zatd-dist: *.go ../go.mod ../go.sum
	GOOS=linux GOARCH=amd64 go build -o zatd-dist .
.PHONY: deploy
deploy: zatd-dist
	rsync -aP zatd-dist horse: && ssh horse 'doas bash -c "install zatd-dist /usr/local/bin/zatd; systemctl restart zatd"'
