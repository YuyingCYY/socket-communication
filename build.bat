@echo off
setlocal

REM 變數
set BINARY_SERVER=server.exe
set BINARY_CLIENT=SocketClient.dll
set BUILD_DIR=build
set CERT_DIR=cert

REM Go 相關變數
set GOCMD=go
set GOBUILD=%GOCMD% build
set GOCLEAN=%GOCMD% clean
set GOTEST=%GOCMD% test
set GOGET=%GOCMD% get
set GOMOD=%GOCMD% mod

REM 版本
set VERSION=1.0.0
for /f "tokens=*" %%a in ('powershell -command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'"') do set BUILD_TIME=%%a

REM 標誌
set LDFLAGS=-ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM 默認目標
if "%1"=="" (
    call :Help
    exit /b
)

if "%1"=="all" (
    call :Clean
    call :Deps
    call :Cert
    call :Build
    call :Test
    exit /b
)

if "%1"=="server" (
    call :BuildServer
    exit /b
)

if "%1"=="client32" (
    call :BuildClient32
    exit /b
)

if "%1"=="client64" (
    call :BuildClient64
    exit /b
)

if "%1"=="clean" (
    call :Clean
    exit /b
)

if "%1"=="test" (
    call :Test
    exit /b
)

if "%1"=="deps" (
    call :Deps
    exit /b
)

if "%1"=="cert" (
    call :Cert
    exit /b
)

if "%1"=="run-server" (
    call :RunServer
    exit /b
)

if "%1"=="help" (
    call :Help
    exit /b
)

:BuildServer
    if not exist "%BUILD_DIR%\server" (
        mkdir "%BUILD_DIR%\server"
    )
    %GOBUILD% %LDFLAGS% -o %BUILD_DIR%\server\%BINARY_SERVER% ./cmd/server
    goto :eof

:BuildClient32
    if not exist "%BUILD_DIR%\client\386" (
        mkdir "%BUILD_DIR%\client\386"
    )
    go env -w GOARCH=386
    %GOBUILD% -buildmode=c-shared -o %BUILD_DIR%\client\386\%BINARY_CLIENT% ./cmd/client
    goto :eof

:BuildClient64
    if not exist "%BUILD_DIR%\client\amd64" (
        mkdir "%BUILD_DIR%\client\amd64"
    )
    go env -w GOARCH=amd64
    %GOBUILD% -buildmode=c-shared -o %BUILD_DIR%\client\amd64\%BINARY_CLIENT% ./cmd/client
    goto :eof

:Clean
    %GOCLEAN%
    if exist "%BUILD_DIR%" (
        rmdir /s /q "%BUILD_DIR%"
    )
    if exist "%CERT_DIR%" (
        del /q "%CERT_DIR%\*.pem"
        del /q "%CERT_DIR%\*.key"
        del /q "%CERT_DIR%\*.crt"
    )
    goto :eof

:Test
    %GOTEST% -v ./...
    goto :eof

:Deps
    %GOMOD% download
    %GOMOD% verify
    goto :eof

:Cert
    if not exist "%CERT_DIR%" (
        mkdir "%CERT_DIR%"
    )
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout "%CERT_DIR%\server.key" -out "%CERT_DIR%\server.crt" -subj "/C=TW/ST=Taiwan/L=Taipei/O=Supreme/CN=localhost"
    goto :eof

:RunServer
    call :BuildServer
    start "" "%BUILD_DIR%\server\%BINARY_SERVER%"
    goto :eof

:Help
    echo commands:
    echo all        - 清理、安裝依賴、生成證書、建構並測試
    echo server     - 建構服務器
    echo client32   - 建構x86客戶端
	echo client64   - 建構x64客戶端
    echo clean      - 清理建構檔案
    echo test       - 執行測試
    echo deps       - 安裝依賴
    echo cert       - 生成 SSL 證書
    echo run-server - 建構並執行服務器
    goto :eof

endlocal