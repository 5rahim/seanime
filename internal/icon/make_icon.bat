@ECHO OFF

IF "%GOPATH%"=="" GOTO ASKGOPATH
:CHECKGOARRAY
IF NOT EXIST %GOPATH%\bin\2goarray.exe GOTO INSTALL
:POSTINSTALL
IF "%1"=="" GOTO NOICO
IF NOT EXIST %1 GOTO BADFILE
ECHO Creating iconwin.go
ECHO //+build windows > iconwin.go
ECHO. >> iconwin.go
TYPE %1 | %GOPATH%\bin\2goarray Data icon >> iconwin.go
GOTO DONE

:CREATEFAIL
ECHO Unable to create output file
GOTO DONE

:INSTALL
ECHO Installing 2goarray...
go install github.com/cratonica/2goarray@latest
IF ERRORLEVEL 1 GOTO GETFAIL
GOTO POSTINSTALL

:GETFAIL
ECHO Failure running go get github.com/cratonica/2goarray.  Ensure that go and git are in PATH
GOTO DONE

:ASKGOPATH
ECHO GOPATH environment variable not set
SET /P GOPATH="Enter GOPATH: "
IF "%GOPATH%"=="" GOTO DONE
ECHO Using GOPATH: %GOPATH%
GOTO CHECKGOARRAY

:NOICO
ECHO Please specify a .ico file
GOTO DONE

:BADFILE
ECHO %1 is not a valid file
GOTO DONE

:DONE

