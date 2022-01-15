zatd: *.go
	GOOS=linux GOARCH=amd64 go build -o zatd .
.PHONY: deploy
deploy: zatd
	rsync -aP zatd ident.me: && ssh ident.me 'doas bash -c "install zatd /usr/local/bin/zatd; systemctl restart zatd"'
