zat: server/main.go
	GOOS=linux GOARCH=amd64 go build -o zat ./server
.PHONY: deploy
deploy: zat
	scp zat ident.me: && ssh ident.me 'sudo bash -c "install zat /usr/local/bin/; systemctl restart zat"'
