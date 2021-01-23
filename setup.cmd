rem @echo off

set ZH_HOME=%1
set SCRIPTDIR=%~dp0

if "%ZH_HOME%" == "" (
	echo setup.cmd targetfolder\
	exit /B 1
)

cd %SCRIPTDIR%
git pull

set FILE=selector.exe
set SOURCE=%SCRIPTDIR%%FILE%
set TARGET=%ZH_HOME%%FILE%

rem go run github.com/rakyll/statik -f -src=static
rem norton internet security ëŒçÙ
if not exist statik.exe go build -o statik.exe github.com/rakyll/statik

statik -f -src=static
go build -o %SOURCE% main.go

if exist %TARGET%.old del /F %TARGET%.old
if exist %TARGET% ren %TARGET% %FILE%.old
copy %SOURCE% %TARGET%

exit /B 0
