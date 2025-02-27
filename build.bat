@echo off
setlocal

REM �ܼ�
set BINARY_SERVER=server.exe
set BINARY_CLIENT=client.dll
set BUILD_DIR=build
set CERT_DIR=cert

REM Go �����ܼ�
set GOCMD=go
set GOBUILD=%GOCMD% build
set GOCLEAN=%GOCMD% clean
set GOTEST=%GOCMD% test
set GOGET=%GOCMD% get
set GOMOD=%GOCMD% mod

REM ����
set VERSION=1.0.0
for /f "tokens=*" %%a in ('powershell -command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'"') do set BUILD_TIME=%%a

REM �лx
set LDFLAGS=-ldflags "-X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME%"

REM �q�{�ؼ�
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

if "%1"=="client" (
    call :BuildClient
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
    if not exist "%BUILD_DIR%" (
        mkdir "%BUILD_DIR%\server"
        mkdir "%BUILD_DIR%\client"
    )
    %GOBUILD% %LDFLAGS% -o %BUILD_DIR%\server\%BINARY_SERVER% ./cmd/server
    goto :eof

:BuildClient
    if not exist "%BUILD_DIR%" (
        mkdir "%BUILD_DIR%\server"
        mkdir "%BUILD_DIR%\client"
    )
    %GOBUILD% -buildmode=c-shared -o %BUILD_DIR%\client\%BINARY_CLIENT% ./cmd/client
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
    echo all        - �M�z�B�w�˨̿�B�ͦ��ҮѡB�غc�ô���
    echo server     - �غc�A�Ⱦ�
    echo client     - �غc�Ȥ��
    echo clean      - �M�z�غc�ɮ�
    echo test       - �������
    echo deps       - �w�˨̿�
    echo cert       - �ͦ� SSL �Ү�
    echo run-server - �غc�ð���A�Ⱦ�
    goto :eof

endlocal