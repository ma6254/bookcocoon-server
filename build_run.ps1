go version
go mod tidy
go mod vendor
.\build.ps1

del ".\release\data.db"
.\run.ps1