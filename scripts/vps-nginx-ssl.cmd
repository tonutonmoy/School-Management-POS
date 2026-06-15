@echo off
REM Run VPS first-time setup (nginx + SSL) via SSH
REM Edit VPS_HOST and VPS_USER below before running.

setlocal
set "VPS_USER=root"
set "VPS_HOST=YOUR_VPS_IP"
set "APP_DIR=/var/www/school-management-pos"
set "GIT_REPO=git@github.com:YOUR_USER/School-Management-POS.git"

if "%VPS_HOST%"=="YOUR_VPS_IP" (
  echo Edit scripts\vps-nginx-ssl.cmd — set VPS_HOST and GIT_REPO
  exit /b 1
)

echo Connecting to %VPS_USER%@%VPS_HOST% ...
echo.
echo DNS must point to VPS before SSL:
echo   school-management.tonusoft.com
echo   school-management-api.tonusoft.com
echo.

ssh -t %VPS_USER%@%VPS_HOST% "export APP_DIR=%APP_DIR% GIT_REPO=%GIT_REPO% && cd %APP_DIR% 2>/dev/null || git clone %GIT_REPO% %APP_DIR% && cd %APP_DIR% && git pull && chmod +x scripts/vps-nginx-ssl.sh && sudo APP_DIR=%APP_DIR% GIT_REPO=%GIT_REPO% ./scripts/vps-nginx-ssl.sh"

endlocal
