# httpPassword
To keep it simple all the files are in package main.  The command file is "httpPassword.go".   To test the handlers there is a test file "httpPassword_test.go".

Valid Urls:

curl —data "password=angryMonkey" http://localhost:8042/hash

curl http://localhost:8042/hash/1

curl http://localhost:8042/stats

curl http://localhost:8042/close

### Installation
go get github.com/depas98/httpPassword

