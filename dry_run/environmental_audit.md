# Environmental Audit (refreshed)

## System
- OS: Ubuntu 22.04
- Kernel: 5.15.0-151-generic

## Python
- Python: 3.10.12

## Virtualenv
- Venv present: yes
- Venv Python: 3.10.12

## Declared deps (pyproject.toml)
dependencies = [
    "websockets>=11.0.3",
    "aiohttp>=3.8.0",
    "PyYAML>=6.0",
    "psutil>=5.9.0",
    # STOP: DRY-02 Minimal fix: declare missing JWT runtime dependency used by src/security/jwt_handler.py
    "PyJWT>=2.4.0",
    # STOP: ENV-AUDIT Minimal fix: declare missing bcrypt runtime dependency used by src/security/api_key_handler.py
    "bcrypt>=4.0.1",
]

## Installed in venv (pip freeze)
aiohappyeyeballs==2.6.1
aiohttp==3.12.15
aiosignal==1.4.0
async-timeout==5.0.1
attrs==25.3.0
backports.asyncio.runner==1.2.0
bcrypt==4.3.0
black==25.1.0
click==8.2.1
coverage==7.10.2
exceptiongroup==1.3.0
flake8==7.3.0
frozenlist==1.7.0
idna==3.10
iniconfig==2.1.0
mccabe==0.7.0
mediamtx-camera-service @ file:///home/dts/CameraRecorder/mediamtx-camera-service
multidict==6.6.3
mypy==1.17.1
mypy_extensions==1.1.0
packaging==25.0
pathspec==0.12.1
platformdirs==4.3.8
pluggy==1.6.0
propcache==0.3.2
psutil==7.0.0
pycodestyle==2.14.0
pyflakes==3.4.0
Pygments==2.19.2
PyJWT==2.10.1
pytest==8.4.1
pytest-asyncio==1.1.0
pytest-cov==6.2.1
PyYAML==6.0.2
tomli==2.2.1
typing_extensions==4.14.1
websockets==15.0.1
yarl==1.20.1

## Import checks (venv)
OK jwt 2.10.1
OK bcrypt 4.3.0
OK websockets 15.0.1
OK aiohttp 3.12.15
OK yaml 6.0.2
OK psutil 7.0.0
