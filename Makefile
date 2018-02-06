DBNAME:=kyuko
TESTDB:=kyuko_dev
ENV:=development

setup:
	go get github.com/rubenv/sql-migrate/...
	go get github.com/Sirupsen/logrus
	go get -u github.com/golang/dep/cmd/dep
	dep get

build:
	go build -o cmd/main cmd/main.go

run:
	./src/cmd/main

test:
	go test -v ./src/... | grep --color=auto -A 3 -e FAIL || echo "PASS ALL TEST"

migrate/init:
	#sudo service mysql restart
	sudo mysql.server restart
	mysql -u root -h localhost --protocol tcp -e "create database \`$(DBNAME)\`" -p

migrate/up:
	sql-migrate up -env="production"

migrate/down:
	sql-migrate down -env="production"

migrate/status:
	sql-migrate status

docker:
	docker build -t kyuko .
