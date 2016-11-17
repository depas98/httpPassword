# httpPassword
To keep it simple all the files are in package main.  The command file is "httpPassword.go".   To test the handlers there is a test file "httpPassword_test.go".

Valid Urls
curl —data "password=angryMonkey" http://localhost:8042/hash
>> 1
curl http://localhost:8042/hash/1
>> "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
curl http://localhost:8042/Stats
>> {“total”:1,“average”:5}


### Installation
go get github.com/depas98/httpPassword

