set "params=%*"
cd /d "%~dp0" && ( if exist "%temp%\getadmin.vbs" del "%temp%\getadmin.vbs" ) && fsutil dirty query %systemdrive% 1>nul 2>nul || (  echo Set UAC = CreateObject^("Shell.Application"^) : UAC.ShellExecute "cmd.exe", "/k cd ""%~sdp0"" && %~s0 %params%", "", "runas", 1 >> "%temp%\getadmin.vbs" && "%temp%\getadmin.vbs" && exit /B )

IF %1==0 GOTO 1
IF %1==1 GOTO 1
IF %1==2 GOTO 2
IF %1==3 GOTO 3

echo Status of service %Name%:
sc query %Name%
echo
echo Please choose what to do:
echo 1. Register and start the service
echo 2. Set your media folder for treeview
echo 3. Stop and unregister the service
echo *. Exit
set /p choce=Your choice:
IF "%choice%" == "1" GOTO 1
IF "%choice%" == "2" GOTO 2
IF "%choice%" == "3" GOTO 3
GOTO endend

:1
echo Registering and starting the service %Name%
sc create %Name% binpath= "%CD%\%Name%.exe -i" start= auto DisplayName= "%Name%"
IF NOT %errorlevel% == 0 GOTO end
sc description %Name% "%Name% for ForkPlayer"
net start %Name%
IF NOT %errorlevel% == 0 GOTO end
echo OK
IF NOT %1 == 0 GOTO end

:2
echo Choose your media folder...
FOR /F "delims=" %%i IN ( 'powershell -Command "(new-object -COM 'Shell.Application').BrowseForFolder(0,'Please choose your media folder:',1,0).self.path"' ) DO set media=%%i
IF NOT "%media%" == "" (
    IF EXIST treeview rmdir treeview
    mklink /D treeview "%media%"
    IF %errorlevel% == 0 echo OK
) ELSE (
    echo Canceled!
)
GOTO end

:3
echo Stopping and unregistering the service %Name%
net stop %Name%
sc delete %Name%
IF %errorlevel% == 0 echo OK

:end
pause
:endend
endlocal