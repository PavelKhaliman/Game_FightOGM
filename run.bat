@echo off
setlocal
cd /d "%~dp0"
if not exist "dist\FightOGM.exe" call "%~dp0build.bat"
if not exist "dist\FightOGM.exe" exit /b 1
start "FightOGM" /D "%~dp0dist" "%~dp0dist\FightOGM.exe"
