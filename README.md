# httpPassword
To keep it simple all the files are in package main.  The command file is "httpPassword.go".  To shutdown the HTTP sever use the link: http://localhost:8042/close.  Run "httpPassword_test.go" to test the handlers.

Valid Urls:

curl —data "password=angryMonkey" http://localhost:8042/hash

curl http://localhost:8042/hash/1

curl http://localhost:8042/stats

curl http://localhost:8042/close

### Installation
go get github.com/depas98/httpPassword

