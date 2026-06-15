@echo off
REM Git pull + Docker build + run on VPS
REM Edit VPS_HOST and VPS_USER below before running.

setlocal
set "VPS_USER=root"
set "VPS_HOST=YOUR_VPS_IP"
set "APP_DIR=/var/www/school-management-pos"
set "BRANCH=main"

if "%VPS_HOST%"=="YOUR_VPS_IP" (
  echo Edit scripts\deploy.cmd — set VPS_HOST to your VPS IP
  exit /b 1
)

echo Deploying to %VPS_USER%@%VPS_HOST% ...
ssh -t %VPS_USER%@%VPS_HOST% "export APP_DIR=%APP_DIR% BRANCH=%BRANCH% && cd %APP_DIR% && chmod +x scripts/deploy.sh && ./scripts/deploy.sh"

endlocal
