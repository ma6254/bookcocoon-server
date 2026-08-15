swag init --parseDependency

$buildTime=$(Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ')
$buildVersion=$(git describe --tags --always)

echo "Build Time: $buildTime"

go build -v -o release/app.exe -ldflags "-X github.com/ma6254/bookcocoon-server/build.BuildTime=$buildTime -X github.com/ma6254/bookcocoon-server/build.BuildVersion=$buildVersion"
