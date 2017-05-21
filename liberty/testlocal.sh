#! /bin/bash
go build . && sudo ACME_EMAIL=graham.anderson@gmail.com ACME_CACHE=/etc/liberty/ APP_ENV=prod LIBERTY_USER=gandalf LIBERTY_PASS=thegrey ./liberty serve
