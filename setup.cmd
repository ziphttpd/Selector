@echo off

set TARGET=%1
set BASE=%~dp0

if "%TARGET%" == "" (
	echo setup.cmd targetfolder\
	exit /B 1
)

cd %BASE%
git pull

set EXEID=selector
set BUILDEXE=%BASE%%EXEID%.exe
set TARGETEXE=%TARGET%%EXEID%.exe

rem go run github.com/rakyll/statik -f -src=static
rem norton internet security ëŒçÙ
if not exist statik.exe go build -o statik.exe github.com/rakyll/statik

statik -f -src=static
go build -o %BUILDEXE% main.go

if exist %TARGETEXE%.old del /F %TARGETEXE%.old
if exist %TARGETEXE% ren %TARGETEXE% %TARGETEXE%.old
copy %BUILDEXE% %TARGETEXE%

exit /B 0
