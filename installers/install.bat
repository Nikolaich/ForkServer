set "params=%*"
cd /d "%~dp0" && ( if exist "%temp%\getadmin.vbs" del "%temp%\getadmin.vbs" ) && fsutil dirty query %systemdrive% 1>nul 2>nul || (  echo Set UAC = CreateObject^("Shell.Application"^) : UAC.ShellExecute "cmd.exe", "/k cd ""%~sdp0"" && %~s0 %params%", "", "runas", 1 >> "%temp%\getadmin.vbs" && "%temp%\getadmin.vbs" && exit /B )

reg Query "HKLM\Hardware\Description\System\CentralProcessor\0" | find /i "x86" > NUL && set arch=386 || set arch=amd64
set dir="C:\Program Files\%Name%"

echo Creating working dirrectory %dir%
IF NOT EXIST %dir% mkdir %dir%
cd %dir%
IF NOT %errorlevel% == 0 GOTO end
echo OK

echo Downloading %Name%.exe
powershell -Command "(New-Object Net.WebClient).DownloadFile('https://github.com/damiva/ForkServer/releases/download/%Vers%/%Name%-windows-%arch%.exe','%Name%.exe')"
IF NOT %errorlevel% == 0 GOTO end
echo OK
echo Downloading manage.bat
powershell -Command "(New-Object Net.WebClient).DownloadFile('https://github.com/damiva/ForkServer/releases/download/%Vers%/manage.bat','manage.bat')"
IF NOT %errorlevel% == 0 GOTO end
echo OK
call manage.bat 0
GOTO endend

:end
pause
:endend
endlocal